package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// newRouter function returns all of the route mappings to the server.
func newRouter(re *urlDataProvider) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range getRoutes(re) {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}
