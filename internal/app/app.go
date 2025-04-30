package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/project/library/internal/entity"
	libraryErrors "github.com/project/library/internal/errors"
	"github.com/project/library/internal/usecase/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/db"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/project/library/config"
	generated "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const SleepTime = time.Second * 3

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.PG.URL)

	if err != nil {
		logger.Error("failed to connect to database", zap.Error(err))
		cancel()
		return
	}

	defer dbPool.Close()

	db.SetupPostgres(dbPool, logger)

	var v *viper.Viper
	if cfg.Outbox.DynamicConfigEnabled {
		v = runViper(cfg, logger)
		defer func() {
			if v != nil {
				v.OnConfigChange(nil)
			}
		}()
	}

	repo := repository.NewPostgresRepository(dbPool)
	outboxRepository := repository.NewOutboxRepository(dbPool)
	transactor := repository.NewTransactor(dbPool, logger)

	runOutbox(ctx, cfg, logger, outboxRepository, transactor)

	useCases := library.New(logger, repo, repo, outboxRepository, transactor)

	ctrl := controller.New(logger, useCases, useCases)

	go runRest(ctx, cfg, logger)
	go runGrpc(cfg, logger, ctrl)

	<-ctx.Done()
	time.Sleep(SleepTime)
}

const DefaultTimeout = 30 * time.Second
const DefaultKeepAlive = 180 * time.Second
const DefaultMaxIdleConns = 100
const DefaultMaxConnsPerHost = 100
const DefaultIdleConnTimeout = 90 * time.Second
const DefaultTLSHandshakeTimeout = 15 * time.Second
const DefaultExpectContinueTimeout = 2 * time.Second

var DefaultMaxIdleConnsPerHost = runtime.GOMAXPROCS(0) + 1

func runOutbox(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) {
	dialer := &net.Dialer{
		Timeout:   DefaultTimeout,
		KeepAlive: DefaultKeepAlive,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          DefaultMaxIdleConns,
		MaxConnsPerHost:       DefaultMaxConnsPerHost,
		IdleConnTimeout:       DefaultIdleConnTimeout,
		TLSHandshakeTimeout:   DefaultTLSHandshakeTimeout,
		ExpectContinueTimeout: DefaultExpectContinueTimeout,
		MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
	}

	client := new(http.Client)
	client.Transport = transport

	globalHandler := globalOutboxHandler(client, cfg.Outbox.BookSendURL, cfg.Outbox.AuthorSendURL, logger)
	outboxService := outbox.New(logger, outboxRepository, globalHandler, cfg, transactor)

	outboxService.Start(
		ctx,
		cfg,
	)
}

func globalOutboxHandler(
	client *http.Client,
	bookURL string,
	authorURL string,
	logger *zap.Logger,

) outbox.GlobalHandler {
	return func(kind repository.OutboxKind) (outbox.KindHandler, error) {
		switch kind {
		case repository.OutboxKindBook:
			return bookOutboxHandler(client, bookURL, logger), nil
		case repository.OutboxKindAuthor:
			return authorOutboxHandler(client, authorURL, logger), nil
		default:
			return nil, fmt.Errorf("unsupported outbox kind: %d", kind)
		}
	}
}

func bookOutboxHandler(client *http.Client, url string, logger *zap.Logger) outbox.KindHandler {
	return func(_ context.Context, data []byte) error {
		book := entity.Book{}
		err := json.Unmarshal(data, &book)

		if err != nil {
			return libraryErrors.NewKindHandlerError("can not deserialize data in book outbox handler: "+err.Error(), -1)
		}

		response, err := client.Post(url, "application/json", strings.NewReader(book.ID))
		defer func() {
			closeErr := response.Body.Close()
			if closeErr != nil {
				logger.Error("failed to close response body: " + closeErr.Error())
			}
		}()

		if err != nil {
			return libraryErrors.NewKindHandlerError("can not send request to book outbox handler: "+err.Error(), -1)
		}

		if response.StatusCode != http.StatusOK {
			return libraryErrors.NewKindHandlerError("can not send request to book outbox handler, status code: "+
				strconv.Itoa(response.StatusCode), response.StatusCode)
		}

		logger.Info("Send book: " + string(data))

		return nil
	}
}

func authorOutboxHandler(client *http.Client, url string, logger *zap.Logger) outbox.KindHandler {
	return func(_ context.Context, data []byte) error {
		author := entity.Author{}
		err := json.Unmarshal(data, &author)

		if err != nil {
			return libraryErrors.NewKindHandlerError("can not deserialize data in author outbox handler: "+err.Error(), -1)
		}

		response, err := client.Post(url, "application/json", strings.NewReader(author.ID))
		defer func() {
			closeErr := response.Body.Close()
			if closeErr != nil {
				logger.Error("failed to close response body: " + closeErr.Error())
			}
		}()

		if err != nil {
			return libraryErrors.NewKindHandlerError("can not send request to author outbox handler: "+err.Error(), -1)
		}

		if response.StatusCode != http.StatusOK {
			return libraryErrors.NewKindHandlerError("can not send request to author outbox handler, status code: "+
				strconv.Itoa(response.StatusCode), response.StatusCode)
		}

		logger.Info("Send author: " + string(data))

		return nil
	}
}

func runRest(ctx context.Context, cfg *config.Config, logger *zap.Logger) {
	mux := grpcruntime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	address := "localhost:" + cfg.GRPC.Port
	err := generated.RegisterLibraryHandlerFromEndpoint(ctx, mux, address, opts)

	if err != nil {
		logger.Error("can not register grpc gateway", zap.Error(err))
		os.Exit(-1)
	}

	gatewayPort := ":" + cfg.GatewayPort
	logger.Info("gateway listening at port", zap.String("port", gatewayPort))

	if err = http.ListenAndServe(gatewayPort, mux); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(cfg *config.Config, logger *zap.Logger, libraryService generated.LibraryServer) {
	port := ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", port)

	if err != nil {
		logger.Error("can not open tcp socket", zap.Error(err))
		os.Exit(-1)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	generated.RegisterLibraryServer(s, libraryService)

	logger.Info("grpc server listening at port", zap.String("port", port))

	if err = s.Serve(lis); err != nil {
		logger.Error("grpc server listen error", zap.Error(err))
	}
}

func runViper(cfg *config.Config, logger *zap.Logger) *viper.Viper {
	logger.Info("run viper")
	v := viper.New()
	v.SetConfigFile("./config/config.yaml")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		logger.Error("Failed to read dynamic outbox config " + cfg.Outbox.DynamicConfigFileName + err.Error())
		return nil
	}

	v.WatchConfig()
	v.OnConfigChange(func(_ fsnotify.Event) {
		newMaxRetries := v.GetInt("outbox.maxRetries")
		if newMaxRetries > 0 {
			cfg.Outbox.Mu.Lock()
			cfg.Outbox.MaxRetries = newMaxRetries
			cfg.Outbox.Mu.Unlock()
			logger.Debug("Updated outbox max retries",
				zap.Int("new_value", newMaxRetries))
		}
	})

	if initialMaxRetries := v.GetInt("outbox.maxRetries"); initialMaxRetries > 0 {
		logger.Debug("Initialize outbox max retries",
			zap.Int("new_value", initialMaxRetries))
		cfg.Outbox.Mu.Lock()
		cfg.Outbox.MaxRetries = initialMaxRetries
		cfg.Outbox.Mu.Unlock()
	}

	return v
}
