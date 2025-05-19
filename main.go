// main.go
package main

import (
	"click-counter/db"
	"click-counter/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	db.InitDB()

	r := mux.NewRouter()
	r.HandleFunc("/counter/{bannerID}", handlers.IncrementCounter).Methods("GET")
	r.HandleFunc("/stats/{bannerID}", handlers.GetStats).Methods("POST")

	log.Println("Starting server on :3000...")
	log.Fatal(http.ListenAndServe(":3000", r))
}
