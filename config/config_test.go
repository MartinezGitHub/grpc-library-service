package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func setupEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GRPC_PORT", "9090")
	t.Setenv("GRPC_GATEWAY_PORT", "8080")
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "postgres")
	t.Setenv("POSTGRES_USER", "go_student")
	t.Setenv("POSTGRES_PASSWORD", "1234567")
	t.Setenv("POSTGRES_MAX_CONN", "10")
}

func TestConfig(t *testing.T) {
	testcases := []struct {
		name          string
		envUnsetKey   string
		expectedError error
	}{
		{"Missing GRPC_PORT", "GRPC_PORT", ErrMissingGRPCPort},
		{"Missing GRPC_GATEWAY_PORT", "GRPC_GATEWAY_PORT", ErrMissingGRPCGatewayPort},
		{"Missing POSTGRES_HOST", "POSTGRES_HOST", ErrMissingPostgresHost},
		{"Missing POSTGRES_PORT", "POSTGRES_PORT", ErrMissingPostgresPort},
		{"Missing POSTGRES_DB", "POSTGRES_DB", ErrMissingPostgresDB},
		{"Missing POSTGRES_USER", "POSTGRES_USER", ErrMissingPostgresUser},
		{"Missing POSTGRES_PASSWORD", "POSTGRES_PASSWORD", ErrMissingPostgresPass},
		{"Missing POSTGRES_MAX_CONN", "POSTGRES_MAX_CONN", ErrMissingPostgresMaxConn},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			setupEnv(t)
			t.Setenv(tc.envUnsetKey, "")

			_, err := NewConfig()
			require.Error(t, err)
			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestNewConfig_Success(t *testing.T) {
	t.Setenv("POSTGRES_DB", "")
	setupEnv(t)

	cfg, err := NewConfig()
	require.NoError(t, err)

	require.Equal(t, "9090", cfg.GRPC.Port)
	require.Equal(t, "8080", cfg.GRPC.GatewayPort)

	require.Equal(t, "localhost", cfg.PG.Host)
	require.Equal(t, "5433", cfg.PG.Port)
	require.Equal(t, "postgres", cfg.PG.DB)
	require.Equal(t, "go_student", cfg.PG.User)
	require.Equal(t, "1234567", cfg.PG.Password)
	require.Equal(t, "10", cfg.PG.MaxConn)

	expectedURL := "postgres://go_student:1234567@localhost:5433/postgres?sslmode=disable"
	require.Equal(t, expectedURL, cfg.PG.URL)
}
