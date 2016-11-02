package api

import (
	"net/http"
	"github.com/gorilla/mux"
)

type Route struct {
	Name 		string
	Method 		string
	Pattern 	string
	HandlerFunc 	http.HandlerFunc
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range GetRoutes() {
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).HandlerFunc(route.HandlerFunc)
	}

	return router
}

func GetRoutes() []Route {
	s := make([]Route, 0)
	s = append(s, Route{
		Name: "generate",
		Method: "POST",
		Pattern: "/generate",
		HandlerFunc: Generate,
	})

	s = append(s, Route{
		Name: "test",
		Method: "POST",
		Pattern: "/test",
		HandlerFunc: Deploy,
	})

	return s
}



