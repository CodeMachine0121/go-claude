package claudecode

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/CodeMachine0121/go-claude/query"
)

func TestPlainAnswer_KnownFailure(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		expectedError error
	}{
		{
			name:          "an unknown conversation",
			text:          "No conversation found with session ID: 00000000-0000-0000-0000-000000000000",
			expectedError: query.ErrConversationNotFound,
		},
		{
			// Recorded from Claude Code 2.1.332 with a missing --mcp-config file.
			name:          "an MCP configuration file that cannot be found",
			text:          "Error: Invalid MCP configuration:\nMCP config file not found: /tmp/nope.json",
			expectedError: query.ErrInvalidMCPConfiguration,
		},
		{
			// Recorded from Claude Code 2.1.332 with an --mcp-config file that is not JSON.
			name:          "an MCP configuration file that cannot be read",
			text:          "Error: Invalid MCP configuration:\nMCP config is not a valid JSON",
			expectedError: query.ErrInvalidMCPConfiguration,
		},
		{name: "anything else explains nothing known", text: "Segmentation fault"},
		{name: "nothing printed", text: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plainAnswer{text: tt.text}.KnownFailure()

			if tt.expectedError == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.expectedError)
			assert.Contains(t, err.Error(), strconv.Quote(tt.text), "what claude code said is kept")
		})
	}
}
