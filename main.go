package main

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"coding-kittens.com/modules/color"
	"coding-kittens.com/modules/livereload"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const CSS_PATH = "./web/styles.css"

func getAccentBaseValue() string {
	// Read the file synchronously
	fileContent, err := os.ReadFile(CSS_PATH)
	if err != nil {
		// Handle error, e.g., log or return an error value
		return ""
	}

	// Convert byte slice to string
	fileContentStr := string(fileContent)

	// Find the line containing --color-accent-base and extract its value
	re := regexp.MustCompile(`--color-accent-base:\s*([^;]+)`)
	match := re.FindStringSubmatch(fileContentStr)

	if len(match) < 2 {
		return ""
	}

	accentBaseValue := strings.TrimSpace(match[1])
	return accentBaseValue
}

type ComputeData func(c *gin.Context) map[string]interface{}

// RouteData represents the data needed for each route
type RouteData struct {
	Title   string
	Content string
	ComputeData ComputeData
}

var routes = map[string]RouteData{
	"/": {
		Title:   "Home Page",
		Content: "about",
		ComputeData: func(c *gin.Context) map[string]interface{} {
			fromDate := time.Date(2015, time.May, 1, 0, 0, 0, 0, time.UTC)
			
			today := time.Now()

			difference := today.Sub(fromDate)

			// Your data calculation logic here
			return map[string]interface{}{
				"yearDiff": int64(difference.Hours()/24/365),
			}
		},
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
			// Retrieve the accent base value
			accentBaseHSL, parseError := color.HextoHSL(getAccentBaseValue())

			if parseError != nil {
				log.Fatal(parseError)
			}

			log.Print("accentBaseHSL",accentBaseHSL)

			t := template.Must(
				template.New(data.Content).ParseFiles("./web/templates/" + data.Content + ".tmpl"),
			)

			var contentBuffer bytes.Buffer

			var templateData map[string]interface{}

			if data.ComputeData != nil {
				templateData = data.ComputeData(c)
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
				"route":             route,
				"template":          template.HTML(contentBuffer.String()),
				"theme":        	 theme,
				"accentHue": 		 accentBaseHSL.H,
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