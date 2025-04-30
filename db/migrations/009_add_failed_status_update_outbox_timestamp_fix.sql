-- +goose Up
CREATE TYPE outbox_status_new AS ENUM ('CREATED', 'IN_PROGRESS', 'SUCCESS', 'FAILED', 'ABANDONED');

ALTER TABLE outbox ALTER COLUMN status DROP DEFAULT;
ALTER TABLE outbox RENAME COLUMN status TO status_old;
ALTER TABLE outbox ADD COLUMN status outbox_status_new NOT NULL DEFAULT 'CREATED';
UPDATE outbox SET status = status_old::text::outbox_status_new;
ALTER TABLE outbox DROP COLUMN status_old;
DROP TYPE outbox_status;

ALTER TYPE outbox_status_new RENAME TO outbox_status;

-- +goose Down
CREATE TYPE outbox_status_old AS ENUM ('CREATED', 'IN_PROGRESS', 'SUCCESS');

ALTER TABLE outbox ALTER COLUMN status DROP DEFAULT;
ALTER TABLE outbox RENAME COLUMN status TO status_new;
ALTER TABLE outbox ADD COLUMN status outbox_status_old NOT NULL DEFAULT 'CREATED';
UPDATE outbox SET status =
                      CASE
                          WHEN status_new = 'FAILED' OR status_new = 'ABANDONED' THEN 'CREATED'::outbox_status_old
                          ELSE status_new::text::outbox_status_old
                          END;
ALTER TABLE outbox DROP COLUMN status_new;
DROP TYPE outbox_status;
