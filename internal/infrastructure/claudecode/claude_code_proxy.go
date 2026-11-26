// Package claudecode runs the Claude Code CLI. It is the only code in the module
// that starts a Claude Code process.
package claudecode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/CodeMachine0121/go-claude/auth"
	"github.com/CodeMachine0121/go-claude/installation"
)

const (
	executableName  = "claude"
	profileVariable = "CLAUDE_CONFIG_DIR"
	waitAfterCancel = time.Second
)

// Settings is everything the composition root decides about running Claude Code.
type Settings struct {
	// CLIPath is a chosen Claude Code executable; empty means search the path.
	CLIPath string
	// ProfileDirectory isolates sign-in and history; empty means the everyday profile.
	ProfileDirectory string
	// Credential is a supplied secret; zero means use whatever Claude Code is signed in with.
	Credential Credential
	// Stdin, Stdout and Stderr are attached to the interactive sign-in.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// ClaudeCodeProxy runs Claude Code with the chosen executable, profile and credential.
// It is safe for concurrent use.
type ClaudeCodeProxy struct {
	settings Settings
}

// NewClaudeCodeProxy builds the proxy.
func NewClaudeCodeProxy(settings Settings) *ClaudeCodeProxy {
	return &ClaudeCodeProxy{settings: settings}
}

// Inspect reports where Claude Code is and which release it is.
func (p *ClaudeCodeProxy) Inspect(ctx context.Context) (installation.InstallationVo, error) {
	command, err := p.command(ctx, "--version")
	if err != nil {
		return installation.InstallationVo{}, err
	}
	output, err := command.Output()
	if err != nil {
		return installation.InstallationVo{}, p.failure(ctx, err)
	}
	version, err := installation.NewVersionVo(string(output))
	if err != nil {
		return installation.InstallationVo{}, err
	}
	return installation.InstallationVo{Path: command.Path, Version: version}, nil
}

// ReadStatus asks Claude Code for the sign-in state. Claude Code exits with a
// failure status when nobody is signed in, so its answer, not its exit status,
// decides the outcome.
func (p *ClaudeCodeProxy) ReadStatus(ctx context.Context) (auth.StatusVo, error) {
	command, err := p.command(ctx, "auth", "status")
	if err != nil {
		return auth.StatusVo{}, err
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()
	var exitErr *exec.ExitError
	if runErr != nil && (ctx.Err() != nil || !errors.As(runErr, &exitErr)) {
		return auth.StatusVo{}, p.failure(ctx, runErr)
	}
	payload, err := newStatusPayload(stdout.Bytes())
	if err != nil {
		said := strings.TrimSpace(stdout.String() + stderr.String())
		return auth.StatusVo{}, &auth.UnreadableStatusError{Output: p.settings.Credential.Redact(said)}
	}
	return payload.ToVo(), nil
}

// SignIn lets the person sign in through Claude Code on the attached terminal.
func (p *ClaudeCodeProxy) SignIn(ctx context.Context, signIn auth.SignInVo) error {
	command, err := p.command(ctx, signInCommand{signIn: signIn}.arguments()...)
	if err != nil {
		return err
	}
	command.Stdin = p.settings.Stdin
	command.Stdout = p.settings.Stdout
	command.Stderr = p.settings.Stderr
	runErr := command.Run()
	var exitErr *exec.ExitError
	if runErr != nil && ctx.Err() == nil && errors.As(runErr, &exitErr) {
		return auth.ErrSignInNotCompleted
	}
	if runErr != nil {
		return p.failure(ctx, runErr)
	}
	return nil
}

// SignOut asks Claude Code to sign out. Claude Code succeeds when already signed out.
func (p *ClaudeCodeProxy) SignOut(ctx context.Context) error {
	command, err := p.command(ctx, "auth", "logout")
	if err != nil {
		return err
	}
	output, err := command.CombinedOutput()
	if err != nil && ctx.Err() == nil {
		said := strings.TrimSpace(string(output))
		return fmt.Errorf("%w: claude code said %q", auth.ErrSignOutFailed, p.settings.Credential.Redact(said))
	}
	if err != nil {
		return p.failure(ctx, err)
	}
	return nil
}

// command prepares a Claude Code invocation with the chosen executable,
// environment and cancellation. It fails with installation.ErrNotInstalled when
// the executable cannot be found; a chosen executable is never replaced by another.
func (p *ClaudeCodeProxy) command(ctx context.Context, arguments ...string) (*exec.Cmd, error) {
	path := p.settings.CLIPath
	if path == "" {
		found, err := exec.LookPath(executableName)
		if err != nil {
			return nil, fmt.Errorf("%w: %q was not found on the search path", installation.ErrNotInstalled, executableName)
		}
		path = found
	} else if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("%w: nothing at %q", installation.ErrNotInstalled, path)
	}
	command := exec.CommandContext(ctx, path, arguments...)
	command.Env = p.environment()
	command.WaitDelay = waitAfterCancel
	return command, nil
}

// environment inherits the program's environment, then applies the profile and
// the credential. A supplied credential replaces both kinds of inherited
// credential so Claude Code cannot pick the other one.
func (p *ClaudeCodeProxy) environment() []string {
	replaced := map[string]bool{}
	if p.settings.ProfileDirectory != "" {
		replaced[profileVariable] = true
	}
	if !p.settings.Credential.IsZero() {
		replaced[oauthTokenVariable] = true
		replaced[apiKeyVariable] = true
	}
	environment := make([]string, 0, len(os.Environ())+2)
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if !replaced[name] {
			environment = append(environment, entry)
		}
	}
	if p.settings.ProfileDirectory != "" {
		environment = append(environment, profileVariable+"="+p.settings.ProfileDirectory)
	}
	if !p.settings.Credential.IsZero() {
		environment = append(environment, p.settings.Credential.variable+"="+p.settings.Credential.secret)
	}
	return environment
}

// failure translates a failed Claude Code run into an error the features
// recognise: given up, not installed, or a failure with secrets hidden.
func (p *ClaudeCodeProxy) failure(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return fmt.Errorf("claude code was given up: %w", ctx.Err())
	}
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", installation.ErrNotInstalled, p.settings.Credential.Redact(err.Error()))
	}
	said := err.Error()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		said += ": " + strings.TrimSpace(string(exitErr.Stderr))
	}
	return errors.New("claude code failed: " + p.settings.Credential.Redact(said))
}
