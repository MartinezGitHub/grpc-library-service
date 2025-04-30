package library

import (
	"context"
	"encoding/json"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/dto"

	"github.com/project/library/internal/usecase/repository"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) RegisterAuthor(ctx context.Context, logger *zap.Logger, authorName string) (*library.RegisterAuthorResponse, error) {
	var author entity.Author
	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var txErr error
		author, txErr = l.authorRepository.CreateAuthor(ctx, logger, entity.Author{
			ID:   uuid.New().String(),
			Name: authorName,
		})

		if txErr != nil {
			return txErr
		}

		serialized, txErr := json.Marshal(author)
		if txErr != nil {
			return txErr
		}

		idempotencyKey := repository.OutboxKindAuthor.String() + "_" + author.ID
		txErr = l.outboxRepository.SendMessage(ctx, idempotencyKey, repository.OutboxKindAuthor, serialized)

		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}

func (l *libraryImpl) ChangeAuthorInfo(ctx context.Context, logger *zap.Logger, authorID string, authorName string) error {
	return l.authorRepository.ChangeAuthorForID(ctx, logger, authorID, authorName)
}

func (l *libraryImpl) GetAuthorByID(ctx context.Context, logger *zap.Logger, authorID string) (entity.Author, error) {
	author, err := l.authorRepository.GetAuthor(ctx, logger, authorID)
	if err != nil {
		return entity.Author{}, err
	}
	return author, nil
}

func (l *libraryImpl) GetAuthorInfo(ctx context.Context, logger *zap.Logger, authorID string) (*library.GetAuthorInfoResponse, error) {
	author, err := l.GetAuthorByID(ctx, logger, authorID)
	if err != nil {
		return nil, err
	}
	return &library.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}

func (l *libraryImpl) StreamBooksForAuthor(ctx context.Context, logger *zap.Logger, authorID string) (<-chan dto.Book, <-chan error) {
	entityCh, errCh := l.booksRepository.StreamBooksByAuthorID(ctx, logger, authorID)

	dtoCh := make(chan dto.Book)
	go func() {
		defer close(dtoCh)
		for book := range entityCh {
			dtoCh <- dto.Book{
				ID:        book.ID,
				Name:      book.Name,
				AuthorIDs: book.AuthorIDs,
			}
		}
	}()

	return dtoCh, errCh
}
