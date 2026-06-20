package main

import (
	"github.com/google/uuid"
)

type Book struct {
	ID              uuid.UUID `json:"id"`
	Author          string    `json:"author"`
	PageCount       int       `json:"page_count"`
	Genre           string    `json:"genre"`
	PublicationDate string    `json:"publication_date"`
}
