package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg := LoadConfig()
	
	repo := NewMockBookRepo()
	handler := NewBookHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /book", handler.CreateBook)
	mux.HandleFunc("GET /book/{id}", handler.GetBook)

	fmt.Printf("Starting server on port %s...\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
