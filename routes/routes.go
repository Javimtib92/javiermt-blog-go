package routes

import "coding-kittens.com/controllers"

type RouteData struct {
	Title   string // metadata title
	Content string // template name to be used. for example for "about.tmpl" Content is equal to "about"
	Controller controllers.ControllerFunc // controller function to send data to the template
}

func GetRoutes() map[string]RouteData {
	return map[string]RouteData{
		"/": {
			Title:      "Home Page",
			Content:    "about",
			Controller: controllers.AboutController,
		},
		"/blog": {
			Title:   "Blog",
			Content: "blog",
		},
	}
}