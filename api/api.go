package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func InstallRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/hello", handleHello).Methods("GET")
}

func JSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, 200, map[string]string{"hello": "world!"})
}
