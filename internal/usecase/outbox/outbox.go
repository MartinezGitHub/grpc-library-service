package outbox

import (
	"context"
	"errors"
	"github.com/project/library/config"
	libraryErrors "github.com/project/library/internal/errors"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
	"sync"
	"time"
)

type GlobalHandler = func(kind repository.OutboxKind) (KindHandler, error)
type KindHandler = func(ctx context.Context, data []byte) error

type Outbox interface {
	Start(ctx context.Context, cfg *config.Config)
	//Start(ctx context.Context, workers int, batchSize int, waitTime time.Duration, inProgressTTL time.Duration, maxRetries int, baseRetryTTL time.Duration)
}

var _ Outbox = (*outboxImpl)(nil)

type outboxImpl struct {
	logger           *zap.Logger
	outboxRepository repository.OutboxRepository
	globalHandler    GlobalHandler
	cfg              *config.Config
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	globalHandler GlobalHandler,
	cfg *config.Config,
	transactor repository.Transactor,
) *outboxImpl {
	return &outboxImpl{
		logger:           logger,
		outboxRepository: outboxRepository,
		globalHandler:    globalHandler,
		cfg:              cfg,
		transactor:       transactor,
	}
}

func (o *outboxImpl) Start(
	ctx context.Context,
	cfg *config.Config,
	// workers int,
	// batchSize int,
	// waitTime time.Duration,
	// inProgressTTL time.Duration,
	// maxRetries int,
	// baseRetryTTL time.Duration,
) {
	wg := new(sync.WaitGroup)

	for workerID := 1; workerID <= cfg.Workers; workerID++ {
		wg.Add(1)
		go o.worker(ctx, wg, cfg)
		//go o.worker(ctx, wg, batchSize, waitTime, inProgressTTL, maxRetries, baseRetryTTL)
	}

	return
}

func (o *outboxImpl) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	cfg *config.Config,
	// batchSize int,
	// waitTime time.Duration,
	// inProgressTTL time.Duration,
	// maxRetries int,
	// baseRetryTTL time.Duration,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			o.logger.Info("worker stopping due to context cancellation")
			return
		default:
		}

		//time.Sleep(waitTime)

		select {
		case <-time.After(cfg.WaitTimeMS):
		case <-ctx.Done():
			o.logger.Info("worker stopping due to context cancellation")
			return
		}

		if !o.cfg.Outbox.Enabled {
			continue
		}

		err := o.transactor.WithTx(ctx, func(ctx context.Context) error {

			if err := ctx.Err(); err != nil {
				return err
			}

			messages, err := o.outboxRepository.GetMessages(ctx, cfg.Outbox.BatchSize, cfg.Outbox.InProgressTTLMS)

			if err != nil {
				o.logger.Error("can not fetch messages from outbox", zap.Error(err))
				return err
			}

			//o.logger.Info("messages fetched", zap.Int("size", len(messages)))
			successKeys := make([]string, 0, len(messages))
			failedKeys := make([]string, 0, len(messages))
			retryKeys := make([]string, 0, len(messages))
			for i := 0; i < len(messages); i++ {
				message := messages[i]
				key := message.IdempotencyKey

				kindHandler, handlerErr := o.globalHandler(message.Kind)

				if handlerErr != nil {
					o.logger.Error("unexpected kind", zap.Error(handlerErr))
					continue
				}

				kindErr := kindHandler(ctx, message.RawData)

				if kindErr != nil {
					var handlerError *libraryErrors.KindHandlerError
					if errors.As(kindErr, &handlerError) {
						status := handlerError.Status()
						o.logger.Error("kind error",
							zap.Error(kindErr),
							zap.Int("status", status),
						)

						if status >= 500 && status < 600 {
							o.logger.Error("Server error (5xx)", zap.Int("status", status))
							retryKeys = append(retryKeys, key)
						} else {
							failedKeys = append(failedKeys, key)
						}
					} else {
						o.logger.Error("kind error (non-HTTP)", zap.Error(kindErr))
					}

					continue

				}

				successKeys = append(successKeys, key)
			}

			err = o.outboxRepository.MarkAsProcessed(ctx, successKeys)
			if err != nil {
				o.logger.Error("mark as processed outbox error", zap.Error(err))
				return err
			}

			maxRetries := getMaxRetries(cfg)

			err = o.outboxRepository.MarkAsFailed(ctx, retryKeys, maxRetries, cfg.BaseRetryTTL)
			if err != nil {
				o.logger.Error("mark as failed outbox error", zap.Error(err))
				return err
			}

			err = o.outboxRepository.MarkAsAbandoned(ctx, failedKeys)
			if err != nil {
				o.logger.Error("mark as abandoned outbox error", zap.Error(err))
				return err
			}

			return nil
		})

		if err != nil {
			o.logger.Error("worker stage error", zap.Error(err))
		}
	}
}

func getMaxRetries(cfg *config.Config) int {
	cfg.Outbox.Mu.RLock()
	defer cfg.Outbox.Mu.RUnlock()
	return cfg.Outbox.MaxRetries
}
