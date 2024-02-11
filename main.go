package main

import (
	"bytes"
	"context"
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"coding-kittens.com/controllers"
	"coding-kittens.com/middlewares"
	"coding-kittens.com/modules/articles"
	"coding-kittens.com/modules/image"
	"coding-kittens.com/modules/livereload"
	"coding-kittens.com/modules/utils"
	"coding-kittens.com/routes"
	ginCompressor "github.com/CAFxX/httpcompression/contrib/gin-gonic/gin"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

//go:embed web/templates/*.tmpl
var templateFiles embed.FS
//go:embed all:web/static
var staticAssets embed.FS
//go:embed all:web/_articles
var articlesFS embed.FS

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	utils.StaticAssets = staticAssets
	image.StaticAssets = staticAssets
	controllers.ArticlesFS = articlesFS
	articles.ArticlesFS = articlesFS

	useHTTPS := flag.Bool("https", false, "start HTTPS server")
	debug := flag.Bool("debug", false, "Enable debug mode")

    flag.Parse()

    if *debug {
        gin.SetMode(gin.DebugMode)
    } else {
        gin.SetMode(gin.ReleaseMode)
    }

	if gin.IsDebugging() {
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

func setupRouter() *gin.Engine {
	router := gin.Default()
	
	compress, err := ginCompressor.DefaultAdapter()
	if err != nil {
		log.Fatal(err)
	}

	router.Use(middlewares.SetContextDataMiddleware())

	router.Use(compress)

	loadTemplates(router)

	for route, data := range routes.GetRoutes() {
		router.GET(route, handleRoute(data, router))
	}

	router.GET("/image", image.ProcessImage)

	router.StaticFS("/static", http.FS(loadFS(staticAssets, "web/static")))
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	return router
}


func loadTemplates(router *gin.Engine) error {
    templateFS, err := template.ParseFS(loadFS(templateFiles, "web/templates"), "*.tmpl")

    if err != nil {
        return err
    }

    router.SetHTMLTemplate(templateFS)

    return nil
}

func loadFS(embed embed.FS, dir string) fs.FS {
	var fileSystem fs.FS

	if gin.IsDebugging() {
		fileSystem = os.DirFS(dir)
	} else {
		fileSystem, _ = fs.Sub(embed, dir)
	}

	return fileSystem
}

func handleRoute(data routes.RouteData, router *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxData := c.MustGet("ContextData").(middlewares.ContextData)

		renderTemplate(c, data, ctxData)
	}
}

func renderTemplate(c *gin.Context, data routes.RouteData, ctxData middlewares.ContextData) {
	t := template.Must(template.New(data.Content).ParseFS(loadFS(templateFiles, "web/templates"), "*.tmpl"))

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

	var rootContentBuffer bytes.Buffer

	template.Must(template.New("root.tmpl").ParseFS(loadFS(templateFiles, "web/templates"), "root.tmpl"))

	renderData := struct {
        LiveReloadEnabled bool
        Title             string
        Description       string
        Route             string
        Template          template.HTML
        AccentHue         float64
    }{
        LiveReloadEnabled: ctxData.LiveReloadEnabled,
        Title:             data.Title,
        Description:       "change to some metadata description, can be overriden on route basis",
        Route:             c.Request.URL.Path,
        Template:          template.HTML(contentBuffer.String()),
        AccentHue:         ctxData.AccentBaseHSL.H,
    }

	err = t.ExecuteTemplate(&rootContentBuffer, "root.tmpl", renderData)

	c.Render(http.StatusOK, render.Data{
		ContentType: "text/html",
		Data:        rootContentBuffer.Bytes(),
	})
}