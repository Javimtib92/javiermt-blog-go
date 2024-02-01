package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"
)

const (
	CSSExtension  = ".css"
	TmplExtension = ".tmpl"
)

var wsMutex sync.Mutex
var wsMessageChannel = make(chan []byte, 100) // Buffered channel for WebSocket messages

type WebSocketMessage struct {
	EventType string `json:"eventType"`
	FileName  string `json:"fileName"`
	FileExt   string `json:"fileExt"`
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.LoadHTMLGlob("web/templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})

	router.Static("/assets", "./web/static/assets")
	router.Static("/css", "./web/static/css")
	router.Static("/fonts", "./web/static/fonts")
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	return router
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add("./web")
	err = watcher.Add("./web/templates")
	if err != nil {
		log.Fatal(err)
	}

	var connections []*websocket.Conn

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case message := <-wsMessageChannel:
				wsMutex.Lock()
				updatedConnections := make([]*websocket.Conn, 0, len(connections))
				for _, conn := range connections {
					go func(conn *websocket.Conn) {
						defer func() {
							if r := recover(); r != nil {
								log.Println("Recovered from panic in WebSocket goroutine:", r)
							}
						}()
						if err := conn.Write(ctx, websocket.MessageText, message); err != nil {
							conn.Close(websocket.StatusInternalError, "Connection error")
						}
					}(conn)
					updatedConnections = append(updatedConnections, conn)
				}
				connections = updatedConnections
				wsMutex.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					ext := filepath.Ext(event.Name)

					if ext == CSSExtension || ext == TmplExtension {
						cmd := exec.CommandContext(ctx, "./tailwindcss", "-i", "./web/styles.css", "-o", "./web/static/css/styles.css")
						err := cmd.Run()
						if err != nil {
							log.Println("Error running tailwindcss:", err)
						}
					}

					log.Println("modified file:", event.Name)

					wsMessage := WebSocketMessage{
						EventType: "FileModified",
						FileName:  event.Name,
						FileExt:   ext,
					}

					message, err := json.Marshal(wsMessage)
					if err != nil {
						log.Println("Error encoding WebSocket message:", err)
						continue
					}

					// Use select to avoid blocking if the channel is full
					select {
					case wsMessageChannel <- message:
					default:
						log.Println("WebSocket message channel is full. Dropping message.")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") != "websocket" {
			http.Error(w, "Not a WebSocket handshake request", http.StatusBadRequest)
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})

		if err != nil {
			log.Println("WebSocket handshake error:", err)
			return
		}

		wsMutex.Lock()
		connections = append(connections, conn)
		wsMutex.Unlock()

		message := []byte("HR: WS connection established")
		if err := conn.Write(ctx, websocket.MessageText, message); err != nil {
			log.Println("Error sending message to the client:", err)
			return
		}
	})

	log.Println("WebSocket server listening on :8081")

	go func() {
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatal(err)
		}
	}()

	r := setupRouter()

	go func() {
		r.Run(":8080")
	}()

	wg.Wait()
}