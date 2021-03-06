package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	api "github.com/bmon/chock/api"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port

	srv := &http.Server{
		Handler: handlers.LoggingHandler(os.Stdout, createRouter()),
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info("Staring server on", addr)
	log.Fatal(srv.ListenAndServe())
}

func createRouter() *mux.Router {
	r := mux.NewRouter()
	api.InstallRoutes(r)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./js/dist/")))

	return r
}
