package claude

import (
	"context"

	"github.com/CodeMachine0121/go-claude/auth"
	"github.com/CodeMachine0121/go-claude/installation"
	"github.com/CodeMachine0121/go-claude/internal/infrastructure/claudecode"
	"github.com/CodeMachine0121/go-claude/query"
)

// Client is the entry point to the library. It is safe for concurrent use.
type Client struct {
	installation installation.InstallationDto
	auth         *auth.UseCase
	query        *query.UseCase
}

// New checks the options, finds Claude Code and verifies it is at least
// installation.MinimumSupportedVersion, then wires every feature.
//
// It fails with ErrConflictingCredentials, ErrInvalidOption,
// installation.ErrNotInstalled or installation.ErrUnsupportedVersion.
func New(ctx context.Context, options ...Option) (*Client, error) {
	settings, err := newConfiguration(options).ToSettings()
	if err != nil {
		return nil, err
	}
	proxy := claudecode.NewClaudeCodeProxy(settings)
	verified, err := installation.NewUseCase(proxy).Verify(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{
		installation: verified,
		auth:         auth.NewUseCase(proxy),
		query:        query.NewUseCase(proxy),
	}, nil
}

// Installation describes the Claude Code CLI in use.
func (c *Client) Installation() installation.InstallationDto {
	return c.installation
}

// Auth reads, starts and ends the Claude Code sign-in.
func (c *Client) Auth() *auth.UseCase {
	return c.auth
}

// Query sends one-shot requests and returns Claude's final reply.
func (c *Client) Query() *query.UseCase {
	return c.query
}
