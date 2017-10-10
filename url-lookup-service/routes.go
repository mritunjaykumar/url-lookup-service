package main

import (
	"net/http"
)

// route type defines the data-structure for a given route.
type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Get all routes
func getRoutes(re *urlDataProvider) []route {
	routes := make([]route, 0)
	routes = append(routes,
		route{"findmalware", "GET",
			"/urlinfo/1", re.GetExistingMalwareUrls})
	return routes
}
