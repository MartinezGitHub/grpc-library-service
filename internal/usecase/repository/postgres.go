package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/project/library/internal/entity"
	"go.uber.org/zap"
)

var _ AuthorRepository = (*postgresImpl)(nil)
var _ BooksRepository = (*postgresImpl)(nil)

type postgresImpl struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *postgresImpl {
	return &postgresImpl{
		db: db,
	}
}

func (p postgresImpl) StreamBooksByAuthorID(ctx context.Context, logger *zap.Logger, authorID string) (<-chan entity.Book, <-chan error) {
	const batchSize = 10
	bookChan := make(chan entity.Book, batchSize)
	errChan := make(chan error, 1)
	go func() {
		defer close(bookChan)
		defer close(errChan)

		const query = `
			SELECT 
                b.id, 
                b.name, 
                b.created_at, 
                b.updated_at,
                array_agg(ab.author_id) AS author_ids
            FROM book b
            JOIN author_book ab ON b.id = ab.book_id
            WHERE b.id IN (
                SELECT book_id
                FROM author_book
                WHERE author_id = $1
            )
            GROUP BY b.id, b.name, b.created_at, b.updated_at
            LIMIT $2 OFFSET $3`
		logger.Debug("query", zap.String("sql", query))
		offset := 0
		for {
			rows, err := p.db.Query(ctx, query, authorID, batchSize, offset)
			if err != nil {
				errChan <- errors.Wrap(err, "failed to query books")
				return
			}
			var hasRows bool
			for rows.Next() {
				hasRows = true
				var book entity.Book
				var authorIDs []string

				if err := rows.Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt, &authorIDs); err != nil {
					errChan <- errors.Wrap(err, "failed to scan book")
					return
				}

				book.AuthorIDs = authorIDs

				select {
				case bookChan <- book:
					logger.Info("placed book " + book.ID + " to chan")
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				}
			}

			if err := rows.Err(); err != nil {
				errChan <- errors.Wrap(err, "error during rows iteration")
				return
			}

			rows.Close()

			if !hasRows {
				break
			}

			offset += batchSize
		}
	}()

	return bookChan, errChan
}

func insertAuthors(ctx context.Context, tx pgx.Tx, bookID string, authorIDs []string) error {
	if len(authorIDs) > 0 {
		const queryAuthorBooks = `
		INSERT INTO author_book (author_id, book_id) 
		VALUES ($1, $2)
		`

		batch := &pgx.Batch{}
		for _, authorID := range authorIDs {
			batch.Queue(queryAuthorBooks, authorID, bookID)
		}

		br := tx.SendBatch(ctx, batch)

		for range authorIDs {
			_, err2 := br.Exec()

			if err2 != nil {
				var pgErr *pgconn.PgError
				if errors.As(err2, &pgErr) && pgErr.Code == "23503" {
					return entity.ErrAuthorNotFound
				}
				return errors.Wrap(err2, "failed to insert author-book link")
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("insert batch closing failed: %w", err)
		}
	}
	return nil
}

func (p postgresImpl) CreateBook(ctx context.Context, logger *zap.Logger, book entity.Book) (entity.Book, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return entity.Book{}, errors.Wrap(err, "failed to begin transaction")
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			logger.Warn("transaction rollback failed", zap.Error(err))
		}
	}(tx, ctx)

	const queryBook = `
	INSERT INTO book (id, name)
	VALUES ($1,$2)
	RETURNING created_at, updated_at
	`
	result := entity.Book{
		ID:        book.ID,
		Name:      book.Name,
		AuthorIDs: book.AuthorIDs,
	}

	err = tx.QueryRow(ctx, queryBook, book.ID, book.Name).Scan(&result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return entity.Book{}, errors.Wrap(err, "failed to insert book")
	}

	err = insertAuthors(ctx, tx, book.ID, book.AuthorIDs)
	if err != nil {
		return entity.Book{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return entity.Book{}, errors.Wrap(err, "failed to commit transaction")
	}

	return result, nil
}

func (p postgresImpl) GetBook(ctx context.Context, logger *zap.Logger, bookID string) (entity.Book, error) {
	const query = `
		SELECT 
			b.id, 
			b.name, 
			b.created_at, 
			b.updated_at,
			array_remove(array_agg(ab.author_id), NULL) AS author_ids
		FROM book b
		LEFT JOIN author_book ab ON b.id = ab.book_id
		WHERE b.id = $1
		GROUP BY b.id, b.name, b.created_at, b.updated_at
	`
	logger.Debug("query", zap.String("sql", query))
	var book entity.Book
	var authorIDs []string

	err := p.db.QueryRow(ctx, query, bookID).
		Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt, &authorIDs)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Book{}, entity.ErrBookNotFound
	}
	if err != nil {
		return entity.Book{}, errors.Wrap(err, "failed to query book")
	}
	book.AuthorIDs = authorIDs
	return book, nil
}

func (p postgresImpl) UpdateBookByID(ctx context.Context, logger *zap.Logger, bookID string, bookName string, authorIDs []string) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			logger.Warn("transaction rollback failed", zap.Error(err))
		}
	}(tx, ctx)

	const queryBooks = `
    UPDATE book
    SET name = $1, updated_at = now()
    WHERE id = $2
    `
	result, err := tx.Exec(ctx, queryBooks, bookName, bookID)
	if err != nil {
		return errors.Wrap(err, "failed to update book name")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return entity.ErrBookNotFound
	}

	const queryDeleteAuthorBooks = `
    DELETE FROM author_book
    WHERE book_id = $1
    `
	_, err = tx.Exec(ctx, queryDeleteAuthorBooks, bookID)
	if err != nil {
		return errors.Wrap(err, "failed to delete old author-book links")
	}

	err = insertAuthors(ctx, tx, bookID, authorIDs)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (p postgresImpl) CreateAuthor(ctx context.Context, _ *zap.Logger, author entity.Author) (entity.Author, error) {
	const query = `
INSERT INTO author (id, name)
VALUES ($1,$2)
`
	result := entity.Author{
		ID:   author.ID,
		Name: author.Name,
	}
	_, err := p.db.Exec(ctx, query, author.ID, author.Name)

	if err != nil {
		return entity.Author{}, errors.Wrap(err, "failed to insert author")
	}

	return result, nil
}

func (p postgresImpl) ChangeAuthorForID(ctx context.Context, _ *zap.Logger, id string, newName string) error {
	const query = `
    UPDATE author
    SET name = $1, updated_at = now()
    WHERE id = $2
    `

	result, err := p.db.Exec(ctx, query, newName, id)
	if err != nil {
		return errors.Wrap(err, "failed to update author")
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return entity.ErrAuthorNotFound
	}

	return nil
}

func (p postgresImpl) GetAuthor(ctx context.Context, _ *zap.Logger, authorID string) (entity.Author, error) {
	const query = `
	SELECT id, name
	FROM author
	WHERE id = $1
	`
	var author entity.Author
	err := p.db.QueryRow(ctx, query, authorID).
		Scan(&author.ID, &author.Name)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Author{}, entity.ErrAuthorNotFound
	}

	if err != nil {
		return entity.Author{}, errors.Wrap(err, "failed to query author")
	}
	return author, nil
}
