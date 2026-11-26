package claude

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/CodeMachine0121/go-claude/internal/infrastructure/claudecode"
)

// ErrConflictingCredentials reports that both WithOAuthToken and WithAPIKey were given.
var ErrConflictingCredentials = errors.New("supply only one kind of credential")

// ErrInvalidOption reports an option given an empty or unusable value.
var ErrInvalidOption = errors.New("invalid option")

// Option configures New.
type Option func(*configuration)

// WithCLIPath uses the Claude Code executable at path instead of searching for
// "claude" on the search path. When nothing exists at path, New fails with
// installation.ErrNotInstalled; no other copy is tried.
func WithCLIPath(path string) Option {
	return func(c *configuration) { c.cliPath = &path }
}

// WithProfile keeps the sign-in and conversation history in directory, separate
// from the person's everyday Claude Code sign-in and from every other profile.
// The directory is created by Claude Code on first use.
func WithProfile(directory string) Option {
	return func(c *configuration) { c.profileDirectory = &directory }
}

// WithOAuthToken uses a long-lived Claude subscription pass, such as one made
// with `claude setup-token`, instead of the interactive sign-in. The token never
// appears in errors.
func WithOAuthToken(token string) Option {
	return func(c *configuration) { c.oauthToken = &token }
}

// WithAPIKey uses a developer account key, billed per use, instead of the
// interactive sign-in. The key never appears in errors.
func WithAPIKey(key string) Option {
	return func(c *configuration) { c.apiKey = &key }
}

// WithTerminal attaches the interactive sign-in to the given streams instead of
// the program's standard input, output and error.
func WithTerminal(stdin io.Reader, stdout, stderr io.Writer) Option {
	return func(c *configuration) {
		c.stdin = stdin
		c.stdout = stdout
		c.stderr = stderr
	}
}

// configuration collects the options given to New.
type configuration struct {
	cliPath          *string
	profileDirectory *string
	oauthToken       *string
	apiKey           *string
	stdin            io.Reader
	stdout           io.Writer
	stderr           io.Writer
}

func newConfiguration(options []Option) configuration {
	result := configuration{stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr}
	for _, option := range options {
		option(&result)
	}
	return result
}

// ToSettings validates the options and converts them into proxy settings.
func (c configuration) ToSettings() (claudecode.Settings, error) {
	if c.oauthToken != nil && c.apiKey != nil {
		return claudecode.Settings{}, ErrConflictingCredentials
	}
	settings := claudecode.Settings{Stdin: c.stdin, Stdout: c.stdout, Stderr: c.stderr}
	if c.cliPath != nil {
		if *c.cliPath == "" {
			return claudecode.Settings{}, fmt.Errorf("%w: claude code path is empty", ErrInvalidOption)
		}
		settings.CLIPath = *c.cliPath
	}
	if c.profileDirectory != nil {
		if *c.profileDirectory == "" {
			return claudecode.Settings{}, fmt.Errorf("%w: profile directory is empty", ErrInvalidOption)
		}
		directory, err := filepath.Abs(*c.profileDirectory)
		if err != nil {
			return claudecode.Settings{}, fmt.Errorf("%w: profile directory: %w", ErrInvalidOption, err)
		}
		settings.ProfileDirectory = directory
	}
	if c.oauthToken != nil {
		if *c.oauthToken == "" {
			return claudecode.Settings{}, fmt.Errorf("%w: oauth token is empty", ErrInvalidOption)
		}
		settings.Credential = claudecode.NewOAuthToken(*c.oauthToken)
	}
	if c.apiKey != nil {
		if *c.apiKey == "" {
			return claudecode.Settings{}, fmt.Errorf("%w: api key is empty", ErrInvalidOption)
		}
		settings.Credential = claudecode.NewAPIKey(*c.apiKey)
	}
	return settings, nil
}
