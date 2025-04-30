package dto

import "time"

type Book struct {
	ID        string
	Name      string
	AuthorIDs []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
