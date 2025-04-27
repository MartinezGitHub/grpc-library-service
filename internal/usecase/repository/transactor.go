package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Transactor interface {
	WithTx(ctx context.Context, function func(ctx context.Context) error) error
}

var _ Transactor = (*transactorImpl)(nil)

type transactorImpl struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewTransactor(db *pgxpool.Pool, logger *zap.Logger) Transactor {
	return &transactorImpl{
		db:     db,
		logger: logger,
	}
}

func (t transactorImpl) WithTx(ctx context.Context, function func(ctx context.Context) error) (txErr error) {
	ctxWithTx, err, tx := injectTx(ctx, t.db)
	if err != nil {
		return errors.Wrap(err, "cannot inject transaction")
	}

	defer func() {
		if txErr != nil {
			tx.Rollback(ctxWithTx)
			//if e := tx.Rollback(ctxWithTx); err != nil {
			//	t.logger.Error("cannot rollback transaction", zap.Error(e))
			//}
			return
		}

		tx.Commit(ctxWithTx)
		//if e := tx.Commit(ctxWithTx); e != nil {
		//	txErr = errors.Wrap(e, "failed to commit transaction")
		//	t.logger.Error("cannot commit transaction", zap.Error(e))
		//}
	}()

	if err = function(ctxWithTx); err != nil {
		txErr = errors.Wrap(err, "cannot execute function")
		return txErr
	}

	return nil
}

type txInjector struct{}

var ErrTxNotFound = errors.New("tx not found in context")

func injectTx(ctx context.Context, pool *pgxpool.Pool) (context.Context, error, pgx.Tx) {
	if tx, err := extractTx(ctx); err == nil {
		return ctx, nil, tx
	}

	tx, err := pool.Begin(ctx)

	if err != nil {
		return nil, err, nil
	}

	return context.WithValue(ctx, txInjector{}, tx), nil, tx
}

func extractTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txInjector{}).(pgx.Tx)

	if !ok {
		return nil, ErrTxNotFound
	}

	return tx, nil
}
