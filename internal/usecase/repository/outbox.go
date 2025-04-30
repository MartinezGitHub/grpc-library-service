package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

var _ OutboxRepository = (*outboxRepository)(nil)

type outboxRepository struct {
	db DBPool
}

func NewOutboxRepository(db DBPool) OutboxRepository {
	return &outboxRepository{
		db: db,
	}
}

func (o outboxRepository) SendMessage(ctx context.Context, idempotencyKey string, kind OutboxKind, message []byte) error {
	const query = `
INSERT INTO outbox (idempotency_key, data, status, kind)
VALUES ($1, $2, 'CREATED', $3)
ON CONFLICT (idempotency_key) DO NOTHING;`

	var err error
	if tx, txErr := extractTx(ctx); txErr == nil {
		_, err = tx.Exec(ctx, query, idempotencyKey, message, kind)
	} else {
		_, err = o.db.Exec(ctx, query, idempotencyKey, message, kind)
	}

	if err != nil {
		return err
	}
	return nil
}

func (o outboxRepository) GetMessages(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]OutboxData, error) {
	const query = `
UPDATE outbox 
SET 
    status = 'IN_PROGRESS',
    attempts = attempts + 1
WHERE idempotency_key IN (
    SELECT idempotency_key
    FROM outbox
    WHERE 
        (status = 'CREATED' OR
        (status = 'IN_PROGRESS' AND updated_at < NOW() - $1::INTERVAL) OR 
        (status = 'FAILED' AND next_retry_at < NOW()))
    ORDER BY created_at
    LIMIT $2
    FOR UPDATE SKIP LOCKED
)
RETURNING idempotency_key, data, kind, attempts;
`

	internal := strconv.Itoa(int(inProgressTTL.Milliseconds())) + " ms"

	var (
		err  error
		rows pgx.Rows
	)

	if tx, txErr := extractTx(ctx); txErr == nil {
		rows, err = tx.Query(ctx, query, internal, batchSize)
	} else {
		rows, err = o.db.Query(ctx, query, internal, batchSize)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]OutboxData, 0)

	for rows.Next() {
		var (
			key      string
			rawData  []byte
			kind     OutboxKind
			attempts int
		)

		if err := rows.Scan(&key, &rawData, &kind, &attempts); err != nil {
			return nil, err
		}

		result = append(result, OutboxData{
			IdempotencyKey: key,
			RawData:        rawData,
			Kind:           kind,
			Attempts:       attempts,
		})
	}

	return result, rows.Err()
}

func (o outboxRepository) MarkAsProcessed(ctx context.Context, idempotencyKeys []string) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}

	const query = `
UPDATE outbox
SET status = 'SUCCESS'
WHERE idempotency_key = ANY($1);
`
	var err error
	if tx, txErr := extractTx(ctx); txErr == nil {
		_, err = tx.Exec(ctx, query, idempotencyKeys)
	} else {
		_, err = o.db.Exec(ctx, query, idempotencyKeys)
	}

	if err != nil {
		return err
	}

	return nil
}

func (o outboxRepository) MarkAsAbandoned(ctx context.Context, idempotencyKeys []string) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}

	const query = `
UPDATE outbox
SET status = 'ABANDONED'::outbox_status
WHERE idempotency_key = ANY($1);
`
	var err error
	if tx, txErr := extractTx(ctx); txErr == nil {
		_, err = tx.Exec(ctx, query, idempotencyKeys)
	} else {
		_, err = o.db.Exec(ctx, query, idempotencyKeys)
	}

	if err != nil {
		return err
	}

	return nil
}

func (o outboxRepository) MarkAsFailed(ctx context.Context, idempotencyKeys []string, maxRetries int, baseRetryTTL time.Duration) error {
	const query = `
UPDATE outbox
SET 
    status = CASE 
        WHEN attempts >= $2 THEN 'ABANDONED'::outbox_status
        ELSE 'FAILED'::outbox_status
    END,
    updated_at = NOW(),
    next_retry_at = CASE
        WHEN attempts >= $2 THEN NULL
        ELSE NOW() + ($3 * POWER(2, attempts) * INTERVAL '1 second')
    END
WHERE idempotency_key = ANY($1);`

	var err error
	baseRetrySeconds := baseRetryTTL.Seconds()

	if tx, txErr := extractTx(ctx); txErr == nil {
		_, err = tx.Exec(ctx, query, idempotencyKeys, maxRetries, baseRetrySeconds)
	} else {
		_, err = o.db.Exec(ctx, query, idempotencyKeys, maxRetries, baseRetrySeconds)
	}

	if err != nil {
		return err
	}

	return nil
}
