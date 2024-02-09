package livereload

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"nhooyr.io/websocket"
)

const (
	CSSExtension  = ".css"
	TmplExtension = ".tmpl"
	GoExtension = ".go"
)

var wsMutex sync.Mutex
var wsMessageChannel = make(chan []byte, 100)

type WebSocketMessage struct {
	EventType string `json:"eventType"`
	FileName  string `json:"fileName"`
	FileExt   string `json:"fileExt"`
}

const debounceDelay = 150 // milliseconds

var debounceTimer *time.Timer

// StartLiveReload initializes the live reload functionality.
func StartLiveReload(ctx context.Context) {
	var wg sync.WaitGroup

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// List of directories to watch
    directories := []string{
        "./web",
        "./modules",
		"./controllers",
		"./middlewares",
		"./routes",
        // Add more directories as needed
    }

    for _, dir := range directories {
        err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				err = watcher.Add(path)
				if err != nil {
					log.Println("Error watching dir:", err)
				}
			}
			return nil
		})
        if err != nil {
            log.Println("Error watching parent dir:", err)
        }
    }

	

	log.Print(watcher.WatchList())

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

				if event.Op&fsnotify.Write == fsnotify.Write {
					ext := filepath.Ext(event.Name)

					if ext == CSSExtension || ext == TmplExtension || ext == GoExtension {
						if debounceTimer != nil {
							debounceTimer.Stop()
						}
						debounceTimer = time.AfterFunc(debounceDelay*time.Millisecond, func() {
							cmd := exec.CommandContext(ctx, "./tailwindcss", "-i", "./web/styles.css", "-o", "./web/static/css/styles.css")
							err := cmd.Run()
							if err != nil {
								log.Println("Error running tailwindcss:", err)
							}

							log.Println("CSS file modified:", event.Name)

							wsMessage := WebSocketMessage{
								EventType: "FileModified",
								FileName:  event.Name,
								FileExt:   ext,
							}

							message, err := json.Marshal(wsMessage)
							if err != nil {
								log.Println("Error encoding WebSocket message:", err)
								return
							}

							select {
							case wsMessageChannel <- message:
							default:
								log.Println("WebSocket message channel is full. Dropping message.")
							}
						})
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

	wg.Wait()
}
