package claudecode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CodeMachine0121/go-claude/query"
)

func TestQueryCommand_Arguments(t *testing.T) {
	team := filepath.Join(t.TempDir(), "team") // absolute on every platform
	tests := []struct {
		name     string
		query    query.QueryVo
		expected []string
	}{
		{
			name:     "the request itself is never an argument",
			query:    query.QueryVo{Prompt: "Reply with exactly: OK"},
			expected: []string{"--print", "--output-format", "stream-json", "--verbose"},
		},
		{
			name:     "a round limit of zero tells claude code no limit",
			query:    query.QueryVo{Prompt: "hi", MaxTurns: 0},
			expected: []string{"--print", "--output-format", "stream-json", "--verbose"},
		},
		{
			name:     "a round limit is passed on",
			query:    query.QueryVo{Prompt: "hi", MaxTurns: 5},
			expected: []string{"--print", "--output-format", "stream-json", "--verbose", "--max-turns", "5"},
		},
		{
			name: "model, allowed tools and conversation are passed on",
			query: query.QueryVo{
				Prompt: "hi", Model: "haiku", AllowedTools: []string{"Bash", "Read"}, SessionID: "session-1",
			},
			expected: []string{
				"--print", "--output-format", "stream-json", "--verbose",
				"--model", "haiku", "--allowedTools", "Bash,Read", "--resume", "session-1",
			},
		},
		{
			name: "outside services, forbidden tools and role instructions are passed on",
			query: query.QueryVo{
				Prompt: "hi", DisallowedTools: []string{"Edit", "Bash(git push:*)"},
				MCPConfigurationPaths: []string{filepath.Join(team, "a.json"), filepath.Join(team, "b.json")}, StrictMCPConfiguration: true,
				AppendSystemPrompt: "-You review code.",
			},
			expected: []string{
				"--print", "--output-format", "stream-json", "--verbose",
				"--disallowedTools", "Edit,Bash(git push:*)", "--mcp-config", filepath.Join(team, "a.json"), filepath.Join(team, "b.json"),
				"--strict-mcp-config", "--append-system-prompt=-You review code.",
			},
		},
		{
			name:     "only the listed outside services, with no list, still says so",
			query:    query.QueryVo{Prompt: "hi", StrictMCPConfiguration: true},
			expected: []string{"--print", "--output-format", "stream-json", "--verbose", "--strict-mcp-config"},
		},
		{
			name:  "a relative outside-service list is found from the work folder",
			query: query.QueryVo{Prompt: "hi", WorkingDirectory: filepath.Join(team, "reviewer"), MCPConfigurationPaths: []string{"tools.json"}},
			expected: []string{
				"--print", "--output-format", "stream-json", "--verbose", "--mcp-config", filepath.Join(team, "reviewer", "tools.json"),
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
	team := filepath.Join(t.TempDir(), "team")
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
			name: "the same choices as one-shot requests are passed on",
			query: query.QueryVo{
				MaxTurns: 4, Model: "haiku", AllowedTools: []string{"Read"}, SessionID: "session-1",
				DisallowedTools: []string{"Write"}, MCPConfigurationPaths: []string{filepath.Join(team, "a.json")},
				StrictMCPConfiguration: true, AppendSystemPrompt: "You review code.",
			},
			expected: []string{
				"--print", "--input-format", "stream-json", "--output-format", "stream-json", "--verbose",
				"--max-turns", "4", "--model", "haiku", "--allowedTools", "Read", "--disallowedTools", "Write",
				"--mcp-config", filepath.Join(team, "a.json"), "--strict-mcp-config", "--append-system-prompt=You review code.",
				"--resume", "session-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, conversationCommand{query: tt.query}.arguments())
		})
	}
}

func TestQueryCommand_Check(t *testing.T) {
	folder := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(folder, "tools.json"), []byte(`{"mcpServers":{}}`), 0o600))
	current, err := os.Getwd()
	require.NoError(t, err)
	fromCurrentFolder, err := filepath.Rel(current, filepath.Join(folder, "tools.json"))
	require.NoError(t, err)

	tests := []struct {
		name          string
		query         query.QueryVo
		expectedError error
	}{
		{name: "no list is nothing to check", query: query.QueryVo{}},
		{name: "a list in the work folder", query: query.QueryVo{WorkingDirectory: folder, MCPConfigurationPaths: []string{"tools.json"}}},
		{name: "a relative list without a work folder is found from the current folder", query: query.QueryVo{MCPConfigurationPaths: []string{fromCurrentFolder}}},
		{name: "a list given in full", query: query.QueryVo{MCPConfigurationPaths: []string{filepath.Join(folder, "tools.json")}}},
		{
			name:          "a list that does not exist is refused",
			query:         query.QueryVo{WorkingDirectory: folder, MCPConfigurationPaths: []string{"tools.json", "missing.json"}},
			expectedError: query.ErrInvalidMCPConfiguration,
		},
		{
			name:          "a work folder that does not exist is refused",
			query:         query.QueryVo{WorkingDirectory: filepath.Join(folder, "missing")},
			expectedError: query.ErrWorkFolderMissing,
		},
		{
			name:          "a list is not a work folder",
			query:         query.QueryVo{WorkingDirectory: filepath.Join(folder, "tools.json")},
			expectedError: query.ErrWorkFolderMissing,
		},
		{
			name:          "a folder is not a list",
			query:         query.QueryVo{MCPConfigurationPaths: []string{folder}},
			expectedError: query.ErrInvalidMCPConfiguration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queryCommand{query: tt.query}.check()

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
