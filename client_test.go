package claude

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_RefusesInvalidOptions(t *testing.T) {
	tests := []struct {
		name          string
		options       []Option
		expectedError error
		expectedText  string
	}{
		{
			name:          "both kinds of credential",
			options:       []Option{WithOAuthToken("sk-ant-oat01-SECRET"), WithAPIKey("sk-ant-api03-SECRET")},
			expectedError: ErrConflictingCredentials,
			expectedText:  "supply only one kind of credential",
		},
		{name: "empty claude code path", options: []Option{WithCLIPath("")}, expectedError: ErrInvalidOption, expectedText: "claude code path"},
		{name: "empty profile directory", options: []Option{WithProfile("")}, expectedError: ErrInvalidOption, expectedText: "profile directory"},
		{name: "empty oauth token", options: []Option{WithOAuthToken("")}, expectedError: ErrInvalidOption, expectedText: "oauth token"},
		{name: "empty api key", options: []Option{WithAPIKey("")}, expectedError: ErrInvalidOption, expectedText: "api key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(context.Background(), tt.options...)

			require.ErrorIs(t, err, tt.expectedError)
			assert.Contains(t, err.Error(), tt.expectedText)
			assert.NotContains(t, err.Error(), "SECRET")
			assert.Nil(t, client)
		})
	}
}
