package claudecode

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CodeMachine0121/go-claude/installation"
)

func TestSettings_Environment(t *testing.T) {
	inherited := []string{
		"PATH=/usr/bin",
		"CLAUDE_CONFIG_DIR=/home/person/.claude",
		"ANTHROPIC_API_KEY=sk-ant-api03-inherited",
		"CLAUDE_CODE_OAUTH_TOKEN=sk-ant-oat01-inherited",
	}
	tests := []struct {
		name     string
		settings Settings
		expected []string
	}{
		{
			name:     "no profile and no credential keeps everything inherited",
			settings: Settings{},
			expected: inherited,
		},
		{
			name:     "a profile replaces the inherited profile location",
			settings: Settings{ProfileDirectory: "/tmp/work"},
			expected: []string{
				"PATH=/usr/bin",
				"ANTHROPIC_API_KEY=sk-ant-api03-inherited",
				"CLAUDE_CODE_OAUTH_TOKEN=sk-ant-oat01-inherited",
				"CLAUDE_CONFIG_DIR=/tmp/work",
			},
		},
		{
			name:     "a subscription pass replaces both inherited credentials",
			settings: Settings{Credential: NewOAuthToken("sk-ant-oat01-supplied")},
			expected: []string{
				"PATH=/usr/bin",
				"CLAUDE_CONFIG_DIR=/home/person/.claude",
				"CLAUDE_CODE_OAUTH_TOKEN=sk-ant-oat01-supplied",
			},
		},
		{
			name:     "a developer key replaces both inherited credentials",
			settings: Settings{Credential: NewAPIKey("sk-ant-api03-supplied")},
			expected: []string{
				"PATH=/usr/bin",
				"CLAUDE_CONFIG_DIR=/home/person/.claude",
				"ANTHROPIC_API_KEY=sk-ant-api03-supplied",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.settings.environment(inherited))
		})
	}
}

func TestSettings_ExecutablePath(t *testing.T) {
	directory := t.TempDir()
	name := "claude"
	if runtime.GOOS == "windows" {
		name = "claude.exe"
	}
	existing := filepath.Join(directory, name)
	require.NoError(t, os.WriteFile(existing, nil, 0o755))

	t.Run("a chosen executable that exists is used as-is", func(t *testing.T) {
		path, err := Settings{CLIPath: existing}.executablePath()

		require.NoError(t, err)
		assert.Equal(t, existing, path)
	})

	t.Run("a chosen executable that does not exist is not replaced", func(t *testing.T) {
		t.Setenv("PATH", directory)

		_, err := Settings{CLIPath: filepath.Join(directory, "missing")}.executablePath()

		require.ErrorIs(t, err, installation.ErrNotInstalled)
	})

	t.Run("claude is found on the search path", func(t *testing.T) {
		t.Setenv("PATH", directory)

		path, err := Settings{}.executablePath()

		require.NoError(t, err)
		assert.Equal(t, existing, path)
	})

	t.Run("claude missing from the search path is not installed", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		_, err := Settings{}.executablePath()

		require.ErrorIs(t, err, installation.ErrNotInstalled)
	})
}
