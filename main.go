package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	livereload "coding-kittens.com/modules"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
  
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	liveReloadEnabled, err := strconv.ParseBool(os.Getenv("LIVE_RELOAD_ENABLED"))

	if err != nil {
		liveReloadEnabled = false
	}

	if liveReloadEnabled {
		go livereload.StartLiveReload(ctx)
	}
	

	r := setupRouter()
	r.Run(":8080")
}