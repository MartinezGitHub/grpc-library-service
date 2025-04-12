package repository

import (
	"context"

	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
)

//go:generate ../../../bin/mockgen -source=interfaces.go -destination=../../mocks/repository_mock.go -package=mocks
type (
	AuthorRepository interface {
		CreateAuthor(ctx context.Context, logger *zap.Logger, author entity.Author) (entity.Author, error)
		ChangeAuthorForID(ctx context.Context, logger *zap.Logger, id string, newName string) error
		GetAuthor(ctx context.Context, logger *zap.Logger, authorID string) (entity.Author, error)
	}

	BooksRepository interface {
		CreateBook(ctx context.Context, logger *zap.Logger, book entity.Book) (entity.Book, error)
		GetBook(ctx context.Context, logger *zap.Logger, bookID string) (entity.Book, error)
		UpdateBookByID(ctx context.Context, logger *zap.Logger, bookID string, bookName string, authorIDs []string) error
		StreamBooksByAuthorID(ctx context.Context, logger *zap.Logger, authorID string) (<-chan entity.Book, <-chan error)
	}
)
