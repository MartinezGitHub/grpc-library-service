-- +goose Up
ALTER TABLE outbox
ADD COLUMN attempts INT NOT NULL DEFAULT 0,
ADD COLUMN next_retry_at TIMESTAMP;

-- +goose Down
ALTER TABLE outbox
DROP COLUMN attempts,
DROP COLUMN next_retry_at;