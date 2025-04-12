-- +goose Up
CREATE INDEX index_create_author_name ON author (name);

-- +goose Down
DROP INDEX index_create_author_name