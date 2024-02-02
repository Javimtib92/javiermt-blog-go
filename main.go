package main

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	livereload "coding-kittens.com/modules"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// RouteData represents the data needed for each route
type RouteData struct {
	Title   string
	Content string
}

var routes = map[string]RouteData{
	"/": {
		Title:   "Home Page",
		Content: "about",
	},
	"/blog": {
		Title:   "Blog",
		Content: "blog",
	},
}

func setupRouter(liveReloadEnabled bool) *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.LoadHTMLGlob("web/templates/*")

	for route, data := range routes {
		route := route
		data := data

		router.GET(route, func(c *gin.Context) {
			t := template.Must(
				template.New(data.Content).ParseFiles("./web/templates/" + data.Content + ".tmpl"),
			)

			var contentBuffer bytes.Buffer
			err := t.ExecuteTemplate(&contentBuffer, data.Content, nil)
			if err != nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.HTML(http.StatusOK, "root.tmpl", gin.H{
				"liveReloadEnabled": liveReloadEnabled,
				"Title":             data.Title,
				"description":       "change", // You can customize this as needed
				"route":             route,
				"template":          template.HTML(contentBuffer.String()),
			})
		})
	}

	router.Static("/assets", "./web/static/assets")
	router.Static("/css", "./web/static/css")
	router.Static("/fonts", "./web/static/fonts")
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	return router
}

func main() {
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


	r := setupRouter(liveReloadEnabled)
	r.Run(":8080")
}