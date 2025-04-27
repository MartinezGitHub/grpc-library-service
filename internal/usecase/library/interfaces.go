package library

import (
	"context"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

//go:generate ../../../bin/mockgen  -source=interfaces.go -destination=../../mocks/library_mock.go -package=mocks
type (
	AuthorUseCase interface {
		RegisterAuthor(ctx context.Context, logger *zap.Logger, authorName string) (string, error)
		ChangeAuthorInfo(ctx context.Context, logger *zap.Logger, authorID string, authorName string) error
		StreamBooksForAuthor(ctx context.Context, logger *zap.Logger, authorID string) (<-chan entity.Book, <-chan error)
		GetAuthorInfo(ctx context.Context, logger *zap.Logger, authorID string) (string, error)
		GetAuthorByID(ctx context.Context, logger *zap.Logger, authorID string) (entity.Author, error)
	}

	BooksUseCase interface {
		RegisterBook(ctx context.Context, logger *zap.Logger, name string, authorIDs []string) (entity.Book, error)
		GetBook(ctx context.Context, logger *zap.Logger, bookID string) (entity.Book, error)
		UpdateBook(ctx context.Context, logger *zap.Logger, bookID string, bookName string, authorIDs []string) error
	}
)

var _ AuthorUseCase = (*libraryImpl)(nil)
var _ BooksUseCase = (*libraryImpl)(nil)

type libraryImpl struct {
	logger           *zap.Logger
	authorRepository repository.AuthorRepository
	booksRepository  repository.BooksRepository
	outboxRepository repository.OutboxRepository
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	authorRepository repository.AuthorRepository,
	booksRepository repository.BooksRepository,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) *libraryImpl {
	return &libraryImpl{
		logger:           logger,
		authorRepository: authorRepository,
		booksRepository:  booksRepository,
		outboxRepository: outboxRepository,
		transactor:       transactor,
	}
}
