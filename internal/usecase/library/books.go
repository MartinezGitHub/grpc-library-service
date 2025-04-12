package library

import (
	"context"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) RegisterBook(ctx context.Context, logger *zap.Logger, name string, authorIDs []string) (entity.Book, error) {
	book, err := l.booksRepository.CreateBook(ctx, logger, entity.Book{
		ID:        uuid.New().String(),
		Name:      name,
		AuthorIDs: authorIDs,
	})
	return book, err
}

func (l *libraryImpl) GetBook(ctx context.Context, logger *zap.Logger, bookID string) (entity.Book, error) {
	return l.booksRepository.GetBook(ctx, logger, bookID)
}

func (l *libraryImpl) UpdateBook(ctx context.Context, logger *zap.Logger, bookID string, bookName string, authorIDs []string) error {
	err := l.booksRepository.UpdateBookByID(ctx, logger, bookID, bookName, authorIDs)
	if err != nil {
		return err
	}
	return nil
}
