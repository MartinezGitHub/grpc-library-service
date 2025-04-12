package config

import (
	"fmt"
	"net"
	"os"

	"github.com/pkg/errors"
)

type (
	Config struct {
		GRPC
		PG
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
)

var (
	ErrMissingGRPCPort        = errors.New("GRPC_PORT environment variable is required")
	ErrMissingGRPCGatewayPort = errors.New("GRPC_GATEWAY_PORT environment variable is required")
	ErrMissingPostgresHost    = errors.New("POSTGRES_HOST environment variable is required")
	ErrMissingPostgresPort    = errors.New("POSTGRES_PORT environment variable is required")
	ErrMissingPostgresDB      = errors.New("POSTGRES_DB environment variable is required")
	ErrMissingPostgresUser    = errors.New("POSTGRES_USER environment variable is required")
	ErrMissingPostgresPass    = errors.New("POSTGRES_PASSWORD environment variable is required")
	ErrMissingPostgresMaxConn = errors.New("POSTGRES_MAX_CONN environment variable is required")
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

	return cfg, nil
}
