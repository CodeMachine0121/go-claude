// Package claudecode runs the Claude Code CLI. It is the only code in the module
// that starts a Claude Code process.
package claudecode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/CodeMachine0121/go-claude/auth"
	"github.com/CodeMachine0121/go-claude/installation"
)

const waitAfterCancel = time.Second

// ClaudeCodeProxy runs Claude Code with the chosen executable, profile and credential.
// It is safe for concurrent use.
type ClaudeCodeProxy struct {
	settings Settings
}

// NewClaudeCodeProxy builds the proxy.
func NewClaudeCodeProxy(settings Settings) *ClaudeCodeProxy {
	return &ClaudeCodeProxy{settings: settings}
}

// execution is a Claude Code run that started and ended on its own.
type execution struct {
	path      string
	output    []byte
	said      string
	succeeded bool
}

// Inspect reports where Claude Code is and which release it is.
func (p *ClaudeCodeProxy) Inspect(ctx context.Context) (installation.InstallationVo, error) {
	run, err := p.execute(ctx, false, "--version")
	if err != nil {
		return installation.InstallationVo{}, err
	}
	if !run.succeeded {
		return installation.InstallationVo{}, fmt.Errorf("claude code failed: %s", run.said)
	}
	version, err := installation.NewVersionVo(string(run.output))
	if err != nil {
		return installation.InstallationVo{}, err
	}
	return installation.InstallationVo{Path: run.path, Version: version}, nil
}

// ReadStatus asks Claude Code for the sign-in state. Claude Code exits with a
// failure status when nobody is signed in, so its answer, not its exit status,
// decides the outcome.
func (p *ClaudeCodeProxy) ReadStatus(ctx context.Context) (auth.StatusVo, error) {
	run, err := p.execute(ctx, false, "auth", "status")
	if err != nil {
		return auth.StatusVo{}, err
	}
	payload, err := newStatusPayload(run.output)
	if err != nil {
		return auth.StatusVo{}, &auth.UnreadableStatusError{Output: run.said}
	}
	return payload.ToVo(), nil
}

// SignIn lets the person sign in through Claude Code on the attached terminal.
func (p *ClaudeCodeProxy) SignIn(ctx context.Context, signIn auth.SignInVo) error {
	run, err := p.execute(ctx, true, signInCommand{signIn: signIn}.arguments()...)
	if err != nil {
		return err
	}
	if !run.succeeded {
		return auth.ErrSignInNotCompleted
	}
	return nil
}

// SignOut asks Claude Code to sign out. Claude Code succeeds when already signed out.
func (p *ClaudeCodeProxy) SignOut(ctx context.Context) error {
	run, err := p.execute(ctx, false, "auth", "logout")
	if err != nil {
		return err
	}
	if !run.succeeded {
		return fmt.Errorf("%w: claude code said %q", auth.ErrSignOutFailed, run.said)
	}
	return nil
}

// execute runs Claude Code once. It fails only when the run could not happen or
// was given up: installation.ErrNotInstalled, the context's error, or a start
// failure. A run that ended with a failure status is reported through
// execution.succeeded. Supplied secrets are hidden from every text it returns.
// With terminal set, the run is attached to the configured streams and nothing
// is captured.
func (p *ClaudeCodeProxy) execute(ctx context.Context, terminal bool, arguments ...string) (execution, error) {
	path, err := p.settings.executablePath()
	if err != nil {
		return execution{}, err
	}
	command := exec.CommandContext(ctx, path, arguments...)
	command.Env = p.settings.environment(os.Environ())
	command.WaitDelay = waitAfterCancel
	var stdout, stderr bytes.Buffer
	if terminal {
		command.Stdin, command.Stdout, command.Stderr = p.settings.Stdin, p.settings.Stdout, p.settings.Stderr
	} else {
		command.Stdout, command.Stderr = &stdout, &stderr
	}

	runErr := command.Run()

	if ctx.Err() != nil {
		return execution{}, fmt.Errorf("claude code was given up: %w", ctx.Err())
	}
	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		reason := p.settings.Credential.Redact(runErr.Error())
		if errors.Is(runErr, fs.ErrNotExist) {
			return execution{}, fmt.Errorf("%w: %s", installation.ErrNotInstalled, reason)
		}
		return execution{}, errors.New("claude code could not start: " + reason)
	}
	said := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	return execution{
		path:      path,
		output:    stdout.Bytes(),
		said:      p.settings.Credential.Redact(said),
		succeeded: runErr == nil,
	}, nil
}
