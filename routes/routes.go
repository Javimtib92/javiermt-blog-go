package routes

import "coding-kittens.com/controllers"

type RouteData struct {
	Title   string
	Content string
	Controller controllers.ControllerFunc
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