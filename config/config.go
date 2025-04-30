package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const DefaultMaxRetries = 5
const DefaultBaseRetryTTL = time.Second
const DefaultDynamicConfigFileName = "config.yaml"

type (
	Config struct {
		GRPC
		PG
		Outbox
	}

	GRPC struct {
		Port        string `env:"GRPC_PORT"`
		GatewayPort string `env:"GRPC_GATEWAY_PORT"`
	}

	PG struct {
		URL      string
		Host     string `env:"POSTGRES_HOST"`
		Port     string `env:"POSTGRES_PORT"`
		DB       string `env:"POSTGRES_DB"`
		User     string `env:"POSTGRES_USER"`
		Password string `env:"POSTGRES_PASSWORD"`
		MaxConn  string `env:"POSTGRES_MAX_CONN"`
	}

	Outbox struct {
		Mu                    sync.RWMutex
		Enabled               bool          `env:"OUTBOX_ENABLED"`
		DynamicConfigEnabled  bool          `env:"OUTBOX_DYNAMIC_CONFIG_ENABLED"`
		Workers               int           `env:"OUTBOX_WORKERS"`
		BatchSize             int           `env:"OUTBOX_BATCH_SIZE"`
		WaitTimeMS            time.Duration `env:"OUTBOX_WAIT_TIME_MS"`
		InProgressTTLMS       time.Duration `env:"OUTBOX_IN_PROGRESS_TTL_MS"`
		AuthorSendURL         string        `env:"OUTBOX_AUTHOR_SEND_URL"`
		BookSendURL           string        `env:"OUTBOX_BOOK_SEND_URL"`
		MaxRetries            int
		BaseRetryTTL          time.Duration
		DynamicConfigFileName string
	}
)

var (
	ErrMissingGRPCPort            = errors.New("GRPC_PORT environment variable is required")
	ErrMissingGRPCGatewayPort     = errors.New("GRPC_GATEWAY_PORT environment variable is required")
	ErrMissingPostgresHost        = errors.New("POSTGRES_HOST environment variable is required")
	ErrMissingPostgresPort        = errors.New("POSTGRES_PORT environment variable is required")
	ErrMissingPostgresDB          = errors.New("POSTGRES_DB environment variable is required")
	ErrMissingPostgresUser        = errors.New("POSTGRES_USER environment variable is required")
	ErrMissingPostgresPass        = errors.New("POSTGRES_PASSWORD environment variable is required")
	ErrMissingPostgresMaxConn     = errors.New("POSTGRES_MAX_CONN environment variable is required")
	ErrMissingOutboxWorkers       = errors.New("OUTBOX_WORKERS environment variable is required when outbox is enabled")
	ErrMissingOutboxBatchSize     = errors.New("OUTBOX_BATCH_SIZE environment variable is required when outbox is enabled")
	ErrMissingOutboxWaitTime      = errors.New("OUTBOX_WAIT_TIME_MS environment variable is required when outbox is enabled")
	ErrMissingOutboxInProgressTTL = errors.New("OUTBOX_IN_PROGRESS_TTL_MS environment variable is required when outbox is enabled")
	ErrMissingOutboxBookSendURL   = errors.New("OUTBOX_BOOK_SEND_URL environment variable is required when outbox is enabled")
	ErrMissingOutboxAuthorSendURL = errors.New("OUTBOX_AUTHOR_SEND_URL environment variable is required when outbox is enabled")
)

func NewConfig() (*Config, error) {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		return nil, ErrMissingGRPCPort
	}

	grpcGatewayPort := os.Getenv("GRPC_GATEWAY_PORT")
	if grpcGatewayPort == "" {
		return nil, ErrMissingGRPCGatewayPort
	}

	pgHost := os.Getenv("POSTGRES_HOST")
	if pgHost == "" {
		return nil, ErrMissingPostgresHost
	}

	pgPort := os.Getenv("POSTGRES_PORT")
	if pgPort == "" {
		return nil, ErrMissingPostgresPort
	}

	pgDB := os.Getenv("POSTGRES_DB")
	if pgDB == "" {
		return nil, ErrMissingPostgresDB
	}

	pgUser := os.Getenv("POSTGRES_USER")
	if pgUser == "" {
		return nil, ErrMissingPostgresUser
	}

	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	if pgPassword == "" {
		return nil, ErrMissingPostgresPass
	}

	pgMaxConn := os.Getenv("POSTGRES_MAX_CONN")
	if pgMaxConn == "" {
		return nil, ErrMissingPostgresMaxConn
	}

	cfg := &Config{
		GRPC: GRPC{
			Port:        grpcPort,
			GatewayPort: grpcGatewayPort,
		},
		PG: PG{
			Host:     pgHost,
			Port:     pgPort,
			DB:       pgDB,
			User:     pgUser,
			Password: pgPassword,
			MaxConn:  pgMaxConn,
		},
	}

	hostPort := net.JoinHostPort(cfg.PG.Host, cfg.PG.Port)
	cfg.PG.URL = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.PG.User,
		cfg.PG.Password,
		hostPort,
		cfg.PG.DB,
	)

	var err error
	enabledStr := os.Getenv("OUTBOX_ENABLED")
	if enabledStr == "" {
		cfg.Outbox.Enabled = false
	} else {
		cfg.Outbox.Enabled, err = strconv.ParseBool(enabledStr)
		if err != nil {
			return nil, errors.Wrap(err, "invalid OUTBOX_ENABLED value")
		}
	}

	if !cfg.Outbox.Enabled {
		return cfg, nil
	}

	workersStr := os.Getenv("OUTBOX_WORKERS")
	if workersStr == "" {
		return nil, ErrMissingOutboxWorkers
	}
	cfg.Outbox.Workers, err = parseInt(workersStr)
	if err != nil {
		return nil, errors.Wrap(err, "invalid OUTBOX_WORKERS value")
	}

	batchSizeStr := os.Getenv("OUTBOX_BATCH_SIZE")
	if batchSizeStr == "" {
		return nil, ErrMissingOutboxBatchSize
	}
	cfg.Outbox.BatchSize, err = parseInt(batchSizeStr)
	if err != nil {
		return nil, errors.Wrap(err, "invalid OUTBOX_BATCH_SIZE value")
	}

	waitTimeStr := os.Getenv("OUTBOX_WAIT_TIME_MS")
	if waitTimeStr == "" {
		return nil, ErrMissingOutboxWaitTime
	}
	cfg.Outbox.WaitTimeMS, err = parseTime(waitTimeStr)
	if err != nil {
		return nil, errors.Wrap(err, "invalid OUTBOX_WAIT_TIME_MS value")
	}

	inProgressTTLStr := os.Getenv("OUTBOX_IN_PROGRESS_TTL_MS")
	if inProgressTTLStr == "" {
		return nil, ErrMissingOutboxInProgressTTL
	}
	cfg.Outbox.InProgressTTLMS, err = parseTime(inProgressTTLStr)
	if err != nil {
		return nil, errors.Wrap(err, "invalid OUTBOX_IN_PROGRESS_TTL_MS value")
	}

	bookSendURL := os.Getenv("OUTBOX_BOOK_SEND_URL")
	if bookSendURL == "" {
		return nil, ErrMissingOutboxBookSendURL
	}
	cfg.Outbox.BookSendURL = bookSendURL

	authorSendURL := os.Getenv("OUTBOX_AUTHOR_SEND_URL")
	if authorSendURL == "" {
		return nil, ErrMissingOutboxAuthorSendURL
	}
	cfg.Outbox.AuthorSendURL = authorSendURL

	dynamicConfigEnabled := os.Getenv("OUTBOX_DYNAMIC_CONFIG_ENABLED")
	if strings.Compare(dynamicConfigEnabled, "true") == 0 {
		cfg.Outbox.DynamicConfigEnabled = true
	} else {
		cfg.Outbox.DynamicConfigEnabled = false
	}

	cfg.Outbox.MaxRetries = DefaultMaxRetries
	cfg.Outbox.BaseRetryTTL = DefaultBaseRetryTTL
	cfg.DynamicConfigFileName = DefaultDynamicConfigFileName

	return cfg, nil
}

func parseTime(s string) (time.Duration, error) {
	t, err := parseInt(s)
	if err != nil {
		return time.Duration(0), err
	}
	return time.Duration(t) * time.Millisecond, nil
}

func parseInt(s string) (int, error) {
	str, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(str), nil
}
