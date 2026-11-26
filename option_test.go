package claude

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTerminal_AttachesTheSignInToTheGivenStreams(t *testing.T) {
	stdin := strings.NewReader("")
	var stdout, stderr bytes.Buffer

	settings, err := newConfiguration([]Option{WithTerminal(stdin, &stdout, &stderr)}).ToSettings()

	require.NoError(t, err)
	assert.Same(t, stdin, settings.Stdin)
	assert.Same(t, &stdout, settings.Stdout)
	assert.Same(t, &stderr, settings.Stderr)
}

func TestWithoutTerminal_TheSignInUsesTheProgramsStreams(t *testing.T) {
	settings, err := newConfiguration(nil).ToSettings()

	require.NoError(t, err)
	assert.Same(t, os.Stdin, settings.Stdin)
	assert.Same(t, os.Stdout, settings.Stdout)
	assert.Same(t, os.Stderr, settings.Stderr)
}

func TestWithProfile_UnresolvableRelativeDirectoryIsRefused(t *testing.T) {
	previous, err := os.Getwd()
	require.NoError(t, err)
	removed := t.TempDir()
	require.NoError(t, os.Chdir(removed))
	t.Cleanup(func() { _ = os.Chdir(previous) })
	t.Setenv("PWD", "")
	require.NoError(t, os.Remove(removed))
	if _, err := os.Getwd(); err == nil {
		t.Skip("this platform still resolves a removed working directory")
	}

	_, err = newConfiguration([]Option{WithProfile("profile")}).ToSettings()

	require.ErrorIs(t, err, ErrInvalidOption)
	assert.Contains(t, err.Error(), "profile directory")
}
