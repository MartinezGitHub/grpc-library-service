package library

import (
	"context"
	"encoding/json"

	"github.com/project/library/generated/api/library"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/internal/usecase/repository"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) RegisterBook(ctx context.Context, logger *zap.Logger, name string, authorIDs []string) (*library.AddBookResponse, error) {
	var book entity.Book
	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var txErr error
		book, txErr = l.booksRepository.CreateBook(ctx, logger, entity.Book{
			ID:        uuid.New().String(),
			Name:      name,
			AuthorIDs: authorIDs,
		})

		if txErr != nil {
			return txErr
		}

		serialized, txErr := json.Marshal(book)

		if txErr != nil {
			return txErr
		}

		idempotencyKey := repository.OutboxKindBook.String() + "_" + book.ID
		txErr = l.outboxRepository.SendMessage(ctx, idempotencyKey, repository.OutboxKindBook, serialized)

		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &library.AddBookResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, err
}

func (l *libraryImpl) GetBook(ctx context.Context, logger *zap.Logger, bookID string) (*library.GetBookInfoResponse, error) {
	book, err := l.booksRepository.GetBook(ctx, logger, bookID)

	if err != nil {
		return nil, err
	}
	return &library.GetBookInfoResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}

func (l *libraryImpl) UpdateBook(ctx context.Context, logger *zap.Logger, bookID string, bookName string, authorIDs []string) error {
	err := l.booksRepository.UpdateBookByID(ctx, logger, bookID, bookName, authorIDs)
	if err != nil {
		return err
	}
	return nil
}
