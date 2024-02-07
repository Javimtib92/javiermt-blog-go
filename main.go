package main

import (
	"bytes"
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"coding-kittens.com/middlewares"
	"coding-kittens.com/modules/image"
	"coding-kittens.com/modules/livereload"
	"coding-kittens.com/routes"
	ginCompressor "github.com/CAFxX/httpcompression/contrib/gin-gonic/gin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	router := gin.Default()

	compress, err := ginCompressor.DefaultAdapter()
	if err != nil {
		log.Fatal(err)
	}

	router.Use(middlewares.SetContextDataMiddleware())

	router.Use(compress)

	router.LoadHTMLGlob("web/templates/*")

	for route, data := range routes.GetRoutes() {
		router.GET(route, handleRoute(data))
	}

	router.GET("/image", image.ProcessImage)
	router.StaticFS("/static", gin.Dir("./web/static", true))
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	return router
}

func handleRoute(data routes.RouteData) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxData := c.MustGet("ContextData").(middlewares.ContextData)

		renderTemplate(c, data, ctxData)
	}
}

func renderTemplate(c *gin.Context, data routes.RouteData, ctxData middlewares.ContextData) {
	t := template.Must(template.New(data.Content).ParseGlob("./web/templates/*.tmpl"))

	var contentBuffer bytes.Buffer
	var templateData map[string]interface{}

	if data.Controller != nil {
		templateData = data.Controller(c)
	} else {
		templateData = make(map[string]interface{})
	}

	err := t.ExecuteTemplate(&contentBuffer, data.Content, templateData)

	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	theme, err := c.Cookie("theme")

	// Set Last-Modified header
    lastModified := time.Now().UTC()
    c.Header("Last-Modified", lastModified.Format(http.TimeFormat))

    // Check If-Modified-Since header
    if ims := c.GetHeader("If-Modified-Since"); ims != "" {
        if t, err := time.Parse(http.TimeFormat, ims); err == nil && lastModified.Before(t.Add(1*time.Second)) {
            c.Status(http.StatusNotModified)
            return
        }
    }
	
	c.HTML(http.StatusOK, "root.tmpl", gin.H{
		"LiveReloadEnabled": ctxData.LiveReloadEnabled,
		"Title":             data.Title,
		"Description":       "change to some metadata description, can be overriden on route basis",
		"Route":             c.Request.URL.Path,
		"Template":          template.HTML(contentBuffer.String()),
		"Theme":             theme,
		"AccentHue":         ctxData.AccentBaseHSL.H,
	})
}

func main() {
	// Parse command-line flags
	useHTTPS := flag.Bool("https", false, "start HTTPS server")
	flag.Parse()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	liveReloadEnabled, err := strconv.ParseBool(os.Getenv("LIVE_RELOAD_ENABLED"))
	if err != nil {
		liveReloadEnabled = false
	}

	if liveReloadEnabled {
		go livereload.StartLiveReload(ctx)
	}

	r := setupRouter()

	var addr string
	var port string
	var url string

	if *useHTTPS {
		addr = "coding-kittens.dev.com"
		port = ":443"
		url = addr
		go func() {
			if err := r.RunTLS(":443", "cert.pem", "key.pem"); err != nil {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		}()
	} else {
		addr = "localhost"
		port = ":8080"
		url = addr + port
		go func() {
			if err := r.Run(":8080"); err != nil {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}()
	}

	log.Printf("Server started on %s", url)

	select {}
}