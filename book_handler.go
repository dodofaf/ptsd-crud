package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BookHandler struct {
	repo BookRepo
}

func NewBookHandler(repo BookRepo) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) RegisterRoutes(r chi.Router) {
	r.Get("/book", h.ListBooks)
	r.Get("/book/{id}", h.GetBookByID)
	r.Post("/book", h.CreateBook)
	r.Put("/book/{id}", h.UpdateBook)
	r.Patch("/book/{id}", h.PatchBook)
	r.Delete("/book/{id}", h.DeleteBook)
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	book, err := h.repo.GetBookByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to fetch book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.repo.ListBooks(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateBook(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	book.ID = uuid.New()

	err = h.repo.CreateBook(r.Context(), &book)
	if err != nil {
		http.Error(w, "Failed to create book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	var book Book
	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateBook(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	book.ID = id

	err = h.repo.UpdateBook(r.Context(), &book)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) PatchBook(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	book, err := h.repo.GetBookByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch book", http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(r.Body).Decode(book)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	book.ID = id

	if err := validateBook(book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.repo.UpdateBook(r.Context(), book)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	err = h.repo.DeleteBook(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete book", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateBook(book *Book) error {
	if strings.TrimSpace(book.Author) == "" {
		return errors.New("Author is required")
	}
	if book.PageCount <= 0 {
		return errors.New("Page count must be greater than 0")
	}
	if strings.TrimSpace(book.Genre) == "" {
		return errors.New("Genre is required")
	}
	if strings.TrimSpace(book.PublicationDate) == "" {
		return errors.New("Publication date is required")
	}
	return nil
}
