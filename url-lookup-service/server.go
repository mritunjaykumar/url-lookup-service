package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	args := os.Args[1:]
	//args := []string{"./params.json"}
	re := newURLDataProvider(args)
	defer re.session.Close()

	if re.session == nil {
		log.Fatal("session is nil.")
	}

	// creates a new top level mux.Router.
	router := newRouter(re)

	// configure the router to always run this handler when it couldn't match a request to any other handler
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("%s not found.\n", r.URL)))
	})
	log.Printf("Starting the service on port 8090...")
	log.Fatal(http.ListenAndServe(":8090", router))
}
