package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mockBookRepo struct {
	books map[uuid.UUID]*Book
}

func newMockBookRepo() *mockBookRepo {
	return &mockBookRepo{
		books: make(map[uuid.UUID]*Book),
	}
}

func (m *mockBookRepo) CreateBook(ctx context.Context, book *Book) error {
	m.books[book.ID] = book
	return nil
}

func (m *mockBookRepo) GetBookByID(ctx context.Context, id uuid.UUID) (*Book, error) {
	book, exists := m.books[id]
	if !exists {
		return nil, pgx.ErrNoRows
	}
	return book, nil
}

func (m *mockBookRepo) ListBooks(ctx context.Context) ([]*Book, error) {
	books := []*Book{}
	for _, b := range m.books {
		books = append(books, b)
	}
	return books, nil
}

func (m *mockBookRepo) UpdateBook(ctx context.Context, book *Book) error {
	if _, exists := m.books[book.ID]; !exists {
		return pgx.ErrNoRows
	}
	m.books[book.ID] = book
	return nil
}

func (m *mockBookRepo) DeleteBook(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.books[id]; !exists {
		return pgx.ErrNoRows
	}
	delete(m.books, id)
	return nil
}

func TestCreateHandler_InvalidJSON(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/book", strings.NewReader("invalid json"))
	rec := httptest.NewRecorder()
	handler.CreateBook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateHandler_EmptyAuthor(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	book := &Book{
		ID:              uuid.New(),
		Author:          "",
		PageCount:       300,
		Genre:           "Fiction",
		PublicationDate: "2024-01-01",
	}

	body, _ := json.Marshal(book)
	req := httptest.NewRequest(http.MethodPost, "/book", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateBook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetHandler_Success(t *testing.T) {
	repo := newMockBookRepo()
	id := uuid.New()
	repo.books[id] = &Book{
		ID:              id,
		Author:          "J.K. Rowling",
		PageCount:       400,
		Genre:           "Fantasy",
		PublicationDate: "1997-06-26",
	}

	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/book/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got Book
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if got.ID != id {
		t.Fatalf("Expected id %s, got %s", id, got.ID)
	}
}

func TestGetHandler_NotFound(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/book/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestListHandler_Success(t *testing.T) {
	repo := newMockBookRepo()
	id := uuid.New()
	repo.books[id] = &Book{
		ID:              id,
		Author:          "J.R.R. Tolkien",
		PageCount:       423,
		Genre:           "Fantasy",
		PublicationDate: "1954-07-29",
	}

	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/book", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got []Book
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("Expected 1 book, got %d", len(got))
	}
	if got[0].ID != id {
		t.Fatalf("Expected id %s, got %s", id, got[0].ID)
	}
}

func TestListHandler_Empty(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/book", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got []Book
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Expected 0 books, got %d", len(got))
	}
}

func TestUpdateHandler_Success(t *testing.T) {
	repo := newMockBookRepo()
	id := uuid.New()
	repo.books[id] = &Book{
		ID:              id,
		Author:          "George Orwell",
		PageCount:       328,
		Genre:           "Dystopian",
		PublicationDate: "1949-06-08",
	}

	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	updated := Book{
		Author:          "George Orwell",
		PageCount:       330,
		Genre:           "Sci-Fi",
		PublicationDate: "1949-06-08",
	}
	body, _ := json.Marshal(updated)
	req := httptest.NewRequest(http.MethodPut, "/book/"+id.String(), bytes.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if repo.books[id].PageCount != 330 || repo.books[id].Genre != "Sci-Fi" {
		t.Fatalf("Book not updated: %+v", repo.books[id])
	}
}

func TestUpdateHandler_InvalidUUID(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body, _ := json.Marshal(&Book{Author: "X", PageCount: 300, Genre: "Y", PublicationDate: "2024-01-01"})
	req := httptest.NewRequest(http.MethodPut, "/book/not-a-uuid", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateHandler_NotFound(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body, _ := json.Marshal(&Book{Author: "X", PageCount: 300, Genre: "Y", PublicationDate: "2024-01-01"})
	req := httptest.NewRequest(http.MethodPut, "/book/"+uuid.New().String(), bytes.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdateHandler_InvalidBody(t *testing.T) {
	repo := newMockBookRepo()
	id := uuid.New()
	repo.books[id] = &Book{ID: id, Author: "Fender", PageCount: 300, Genre: "Fiction", PublicationDate: "2020-01-01"}

	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body, _ := json.Marshal(&Book{Author: "", PageCount: 0, Genre: "", PublicationDate: ""})
	req := httptest.NewRequest(http.MethodPut, "/book/"+id.String(), bytes.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestPatchHandler_Success(t *testing.T) {
	repo := newMockBookRepo()
	id := uuid.New()
	repo.books[id] = &Book{
		ID:              id,
		Author:          "Orwell",
		PageCount:       300,
		Genre:           "Dystopian",
		PublicationDate: "2020-01-01",
	}

	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPatch, "/book/"+id.String(), strings.NewReader(`{"page_count": 350}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if repo.books[id].PageCount != 350 {
		t.Fatalf("Expected page_count 350, got %d", repo.books[id].PageCount)
	}
	if repo.books[id].Author != "Orwell" {
		t.Fatalf("Expected author unchanged, got %s", repo.books[id].Author)
	}
}

func TestPatchHandler_NotFound(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPatch, "/book/"+uuid.New().String(), strings.NewReader(`{"page_count": 350}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestPatchHandler_InvalidUUID(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPatch, "/book/not-a-uuid", strings.NewReader(`{"page_count": 350}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestPatchHandler_InvalidBody(t *testing.T) {
	repo := newMockBookRepo()
	id := uuid.New()
	repo.books[id] = &Book{ID: id, Author: "Orwell", PageCount: 300, Genre: "Dystopian", PublicationDate: "2020-01-01"}

	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPatch, "/book/"+id.String(), strings.NewReader(`{"author": ""}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDeleteHandler_Success(t *testing.T) {
	repo := newMockBookRepo()
	id := uuid.New()
	repo.books[id] = &Book{ID: id, Author: "Orwell", PageCount: 300, Genre: "Dystopian", PublicationDate: "2020-01-01"}

	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodDelete, "/book/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	if _, exists := repo.books[id]; exists {
		t.Fatalf("Book was not deleted")
	}
}

func TestDeleteHandler_NotFound(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodDelete, "/book/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestDeleteHandler_InvalidUUID(t *testing.T) {
	repo := newMockBookRepo()
	handler := NewBookHandler(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodDelete, "/book/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
