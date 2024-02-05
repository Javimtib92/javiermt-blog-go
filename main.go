package main

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"coding-kittens.com/middlewares"
	"coding-kittens.com/modules/color"
	"coding-kittens.com/modules/livereload"
	"coding-kittens.com/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter(liveReloadEnabled bool) *gin.Engine {
	router := gin.Default()

	router.Use(middlewares.SetContextDataMiddleware())

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.LoadHTMLGlob("web/templates/*")

	routesMap := routes.GetRoutes()

	for route, data := range routesMap {
		route := route
		data := data

		router.GET(route, handleRoute(data, liveReloadEnabled))
	}

	router.Static("/assets", "./web/static/assets")
	router.Static("/css", "./web/static/css")
	router.Static("/fonts", "./web/static/fonts")
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	return router
}

func handleRoute(data routes.RouteData, liveReloadEnabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the custom context data
		contextData, exists := c.Get("ContextData")
		if !exists {
			log.Fatal("ContextData not set")
			return
		}

		// Type assert the context data to the custom struct
		ctxData, ok := contextData.(middlewares.ContextData)
		if !ok {
			log.Fatal("Failed to type assert ContextData")
			return
		}

		renderTemplate(c, data, ctxData.LiveReloadEnabled, ctxData.AccentBaseHSL)
	}
}

func renderTemplate(c *gin.Context, data routes.RouteData, liveReloadEnabled bool, accentBaseHSL color.HSL) {
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

	c.HTML(http.StatusOK, "root.tmpl", gin.H{
		"liveReloadEnabled": liveReloadEnabled,
		"Title":             data.Title,
		"description":       "change",
		"route":             c.Request.URL.Path,
		"template":          template.HTML(contentBuffer.String()),
		"theme":             theme,
		"accentHue":         accentBaseHSL.H,
	})
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