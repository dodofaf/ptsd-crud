package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookRepo interface {
	GetBookByID(ctx context.Context, id uuid.UUID) (*Book, error)
	ListBooks(ctx context.Context) ([]*Book, error)
	CreateBook(ctx context.Context, book *Book) error
	UpdateBook(ctx context.Context, book *Book) error
	DeleteBook(ctx context.Context, id uuid.UUID) error
}

type bookRepo struct {
	dbpool *pgxpool.Pool
}

func NewBookRepo(dbpool *pgxpool.Pool) BookRepo {
	return &bookRepo{dbpool: dbpool}
}

func (r *bookRepo) GetBookByID(ctx context.Context, id uuid.UUID) (*Book, error) {
	query := `SELECT id, author, page_count, genre, publication_date::text FROM book WHERE id = $1`
	book := &Book{}
	err := r.dbpool.QueryRow(ctx, query, id).Scan(&book.ID, &book.Author, &book.PageCount, &book.Genre, &book.PublicationDate)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (r *bookRepo) CreateBook(ctx context.Context, book *Book) error {
	query := `INSERT INTO book (id, author, page_count, genre, publication_date) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.dbpool.Exec(ctx, query, book.ID, book.Author, book.PageCount, book.Genre, book.PublicationDate)
	return err
}

func (r *bookRepo) ListBooks(ctx context.Context) ([]*Book, error) {
	query := `SELECT id, author, page_count, genre, publication_date::text FROM book ORDER BY author`
	rows, err := r.dbpool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []*Book{}
	for rows.Next() {
		book := &Book{}
		err := rows.Scan(&book.ID, &book.Author, &book.PageCount, &book.Genre, &book.PublicationDate)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, rows.Err()
}

func (r *bookRepo) UpdateBook(ctx context.Context, book *Book) error {
	query := `UPDATE book SET author = $2, page_count = $3, genre = $4, publication_date = $5 WHERE id = $1`
	tag, err := r.dbpool.Exec(ctx, query, book.ID, book.Author, book.PageCount, book.Genre, book.PublicationDate)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *bookRepo) DeleteBook(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM book WHERE id = $1`
	tag, err := r.dbpool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
