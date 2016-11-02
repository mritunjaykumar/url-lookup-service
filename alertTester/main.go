package main

import (
	"log"
	"net/http"
	"fmt"
	"github.com/mritunjaykumar/alertDispatcher/api"
)

func main() {
	fmt.Println("starting alert test service...")
	router := api.NewRouter()
	port := api.GetPortForApi()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}