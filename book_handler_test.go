package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateBook(t *testing.T) {
	repo := NewMockBookRepo()
	handler := NewBookHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /book", handler.CreateBook)

	t.Run("Valid Book", func(t *testing.T) {
		body := []byte(`{"title": "The Hobbit", "author": "J.R.R. Tolkien", "published_year": 1937}`)
		req := httptest.NewRequest(http.MethodPost, "/book", bytes.NewBuffer(body))
		
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %v, got %v", http.StatusCreated, rr.Code)
		}
	})

	t.Run("Missing Title", func(t *testing.T) {
		body := []byte(`{"author": "Unknown", "published_year": 2024}`)
		req := httptest.NewRequest(http.MethodPost, "/book", bytes.NewBuffer(body))
		
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %v, got %v", http.StatusBadRequest, rr.Code)
		}
	})
}

func TestGetBook(t *testing.T) {
	repo := NewMockBookRepo()
	handler := NewBookHandler(repo)

	testBook := Book{
		ID:            "123e4567-e89b-12d3-a456-426614174000",
		Title:         "Dune",
		Author:        "Frank Herbert",
		PublishedYear: 1965,
	}
	repo.Create(testBook)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /book/{id}", handler.GetBook)

	t.Run("Valid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/book/"+testBook.ID, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
	})
}
