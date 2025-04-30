package config

import (
	"testing"
	"time"

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

	t.Setenv("OUTBOX_ENABLED", "true")
	t.Setenv("OUTBOX_WORKERS", "5")
	t.Setenv("OUTBOX_BATCH_SIZE", "100")
	t.Setenv("OUTBOX_WAIT_TIME_MS", "500")
	t.Setenv("OUTBOX_IN_PROGRESS_TTL_MS", "30000")
	t.Setenv("OUTBOX_BOOK_SEND_URL", "http://book-service/api/books")
	t.Setenv("OUTBOX_AUTHOR_SEND_URL", "http://author-service/api/authors")
	t.Setenv("OUTBOX_DYNAMIC_CONFIG_ENABLED", "true")
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

		{"Missing OUTBOX_WORKERS when enabled", "OUTBOX_WORKERS", ErrMissingOutboxWorkers},
		{"Missing OUTBOX_BATCH_SIZE when enabled", "OUTBOX_BATCH_SIZE", ErrMissingOutboxBatchSize},
		{"Missing OUTBOX_WAIT_TIME_MS when enabled", "OUTBOX_WAIT_TIME_MS", ErrMissingOutboxWaitTime},
		{"Missing OUTBOX_IN_PROGRESS_TTL_MS when enabled", "OUTBOX_IN_PROGRESS_TTL_MS", ErrMissingOutboxInProgressTTL},
		{"Missing OUTBOX_BOOK_SEND_URL when enabled", "OUTBOX_BOOK_SEND_URL", ErrMissingOutboxBookSendURL},
		{"Missing OUTBOX_AUTHOR_SEND_URL when enabled", "OUTBOX_AUTHOR_SEND_URL", ErrMissingOutboxAuthorSendURL},
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

	require.True(t, cfg.Outbox.Enabled)
	require.Equal(t, 5, cfg.Outbox.Workers)
	require.Equal(t, 100, cfg.Outbox.BatchSize)
	require.Equal(t, 500*time.Millisecond, cfg.Outbox.WaitTimeMS)
	require.Equal(t, 30000*time.Millisecond, cfg.Outbox.InProgressTTLMS)
	require.Equal(t, "http://book-service/api/books", cfg.Outbox.BookSendURL)
	require.Equal(t, "http://author-service/api/authors", cfg.Outbox.AuthorSendURL)
	require.True(t, cfg.Outbox.DynamicConfigEnabled)
	require.Equal(t, DefaultMaxRetries, cfg.Outbox.MaxRetries)
	require.Equal(t, DefaultBaseRetryTTL, cfg.Outbox.BaseRetryTTL)
	require.Equal(t, DefaultDynamicConfigFileName, cfg.DynamicConfigFileName)
}

func TestOutboxDisabled(t *testing.T) {
	setupEnv(t)
	t.Setenv("OUTBOX_ENABLED", "false")

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.False(t, cfg.Outbox.Enabled)
}

func TestInvalidOutboxValues(t *testing.T) {
	testcases := []struct {
		name        string
		envKey      string
		envValue    string
		expectedErr string
	}{
		{"Invalid OUTBOX_ENABLED", "OUTBOX_ENABLED", "not-a-bool", "invalid OUTBOX_ENABLED value"},
		{"Invalid OUTBOX_WORKERS", "OUTBOX_WORKERS", "not-a-number", "invalid OUTBOX_WORKERS value"},
		{"Invalid OUTBOX_BATCH_SIZE", "OUTBOX_BATCH_SIZE", "not-a-number", "invalid OUTBOX_BATCH_SIZE value"},
		{"Invalid OUTBOX_WAIT_TIME_MS", "OUTBOX_WAIT_TIME_MS", "not-a-number", "invalid OUTBOX_WAIT_TIME_MS value"},
		{"Invalid OUTBOX_IN_PROGRESS_TTL_MS", "OUTBOX_IN_PROGRESS_TTL_MS", "not-a-number", "invalid OUTBOX_IN_PROGRESS_TTL_MS value"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			setupEnv(t)
			t.Setenv(tc.envKey, tc.envValue)

			_, err := NewConfig()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestOutboxDefaultValues(t *testing.T) {
	setupEnv(t)
	t.Setenv("OUTBOX_DYNAMIC_CONFIG_ENABLED", "false")

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.False(t, cfg.Outbox.DynamicConfigEnabled)
	require.Equal(t, DefaultMaxRetries, cfg.Outbox.MaxRetries)
	require.Equal(t, DefaultBaseRetryTTL, cfg.Outbox.BaseRetryTTL)
	require.Equal(t, DefaultDynamicConfigFileName, cfg.DynamicConfigFileName)
}

func TestOutboxEnabledEmpty(t *testing.T) {
	setupEnv(t)
	t.Setenv("OUTBOX_ENABLED", "")

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.False(t, cfg.Outbox.Enabled, "Outbox should be disabled when OUTBOX_ENABLED is empty")
}
