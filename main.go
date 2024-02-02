package main

import (
	"context"
	"net/http"

	livereload "coding-kittens.com/modules"
	"github.com/gin-gonic/gin"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go livereload.StartLiveReload(ctx)

	r := setupRouter()
	r.Run(":8080")
}