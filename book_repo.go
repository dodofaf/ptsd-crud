package main

import (
	"errors"
	"sync"
)

// The interface ensures our handlers don't care if it's a mock or a real DB
type BookRepository interface {
	Create(book Book) error
	GetByID(id string) (Book, error)
}

// In-Memory Mock Implementation
type mockBookRepo struct {
	mu    sync.RWMutex
	books map[string]Book
}

func NewMockBookRepo() BookRepository {
	return &mockBookRepo{
		books: make(map[string]Book),
	}
}

func (m *mockBookRepo) Create(book Book) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.books[book.ID] = book
	return nil
}

func (m *mockBookRepo) GetByID(id string) (Book, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	book, exists := m.books[id]
	if !exists {
		return Book{}, errors.New("book not found")
	}
	return book, nil
}
