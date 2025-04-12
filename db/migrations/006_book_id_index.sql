-- +goose Up
CREATE INDEX index_create_author_books_book_id ON author_book (book_id);

-- +goose Down
DROP INDEX index_create_author_books_book_id