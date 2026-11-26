package claudecode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/CodeMachine0121/go-claude/query"
)

func TestQueryCommand_Arguments(t *testing.T) {
	tests := []struct {
		name     string
		query    query.QueryVo
		expected []string
	}{
		{
			name:     "the request itself is never an argument",
			query:    query.QueryVo{Prompt: "Reply with exactly: OK"},
			expected: []string{"--print", "--output-format", "json"},
		},
		{
			name:     "a round limit of zero tells claude code no limit",
			query:    query.QueryVo{Prompt: "hi", MaxTurns: 0},
			expected: []string{"--print", "--output-format", "json"},
		},
		{
			name:     "a round limit is passed on",
			query:    query.QueryVo{Prompt: "hi", MaxTurns: 5},
			expected: []string{"--print", "--output-format", "json", "--max-turns", "5"},
		},
		{
			name: "model, allowed tools and conversation are passed on",
			query: query.QueryVo{
				Prompt: "hi", Model: "haiku", AllowedTools: []string{"Bash", "Read"}, SessionID: "session-1",
			},
			expected: []string{
				"--print", "--output-format", "json",
				"--model", "haiku", "--allowedTools", "Bash,Read", "--resume", "session-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, queryCommand{query: tt.query}.arguments())
		})
	}
}

func TestConversationCommand_Arguments(t *testing.T) {
	tests := []struct {
		name     string
		query    query.QueryVo
		expected []string
	}{
		{
			name:     "messages and answers are exchanged as json lines",
			query:    query.QueryVo{},
			expected: []string{"--print", "--input-format", "stream-json", "--output-format", "stream-json", "--verbose"},
		},
		{
			name:  "the same choices as one-shot requests are passed on",
			query: query.QueryVo{MaxTurns: 4, Model: "haiku", AllowedTools: []string{"Read"}, SessionID: "session-1"},
			expected: []string{
				"--print", "--input-format", "stream-json", "--output-format", "stream-json", "--verbose",
				"--max-turns", "4", "--model", "haiku", "--allowedTools", "Read", "--resume", "session-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, conversationCommand{query: tt.query}.arguments())
		})
	}
}
