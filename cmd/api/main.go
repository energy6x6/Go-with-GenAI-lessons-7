package main

import (
	"log"
	"net/http"
	"os"

	"homework/internal/app"
)

// @title Book Catalog API
// @version 1.0
// @description Lesson 7: in-memory Book CRUD API. Optional fields are nullable; PUT fully replaces editable fields.
// @BasePath /api/v1
func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, app.NewRouter()); err != nil {
		log.Fatal(err)
	}
}
