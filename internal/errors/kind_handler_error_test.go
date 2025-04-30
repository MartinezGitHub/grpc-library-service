package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKindHandlerError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		message        string
		status         int
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Standard error",
			message:        "Not found",
			status:         404,
			expectedErr:    "Not found status_code: 404",
			expectedStatus: 404,
		},
		{
			name:           "Empty message",
			message:        "",
			status:         500,
			expectedErr:    " status_code: 500",
			expectedStatus: 500,
		},
		{
			name:           "Zero status",
			message:        "Bad request",
			status:         0,
			expectedErr:    "Bad request status_code: 0",
			expectedStatus: 0,
		},
		{
			name:           "Special characters",
			message:        "Error: 测试",
			status:         400,
			expectedErr:    "Error: 测试 status_code: 400",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := NewKindHandlerError(tt.message, tt.status)
			require.NotNil(t, err)
			require.Equal(t, tt.message, err.message)
			require.Equal(t, tt.status, err.status)

			errorStr := err.Error()
			require.Equal(t, tt.expectedErr, errorStr)

			status := err.Status()
			require.Equal(t, tt.expectedStatus, status)

			var errInterface error = err
			require.Equal(t, tt.expectedErr, errInterface.Error())
		})
	}
}
