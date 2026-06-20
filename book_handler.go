package main

import (
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
)

type BookHandler struct {
	repo BookRepository
}

func NewBookHandler(repo BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if book.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	book.ID = uuid.New().String()
	
	if err := h.repo.Create(book); err != nil {
		http.Error(w, "Failed to create book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	book, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}
