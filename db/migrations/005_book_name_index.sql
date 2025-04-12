-- +goose Up
CREATE INDEX index_create_book_name ON book (name);

-- +goose Down
DROP INDEX index_create_book_name