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
	"coding-kittens.com/modules/image"
	"coding-kittens.com/modules/livereload"
	"coding-kittens.com/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.Use(middlewares.SetContextDataMiddleware())

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
	r.Run(":8080")
}