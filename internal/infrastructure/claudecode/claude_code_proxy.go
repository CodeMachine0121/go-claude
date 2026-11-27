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
	"github.com/CodeMachine0121/go-claude/query"
)

const waitAfterCancel = 500 * time.Millisecond

// ClaudeCodeProxy runs Claude Code with the chosen executable, profile and credential.
// It is safe for concurrent use.
type ClaudeCodeProxy struct {
	settings Settings
}

// NewClaudeCodeProxy builds the proxy.
func NewClaudeCodeProxy(settings Settings) *ClaudeCodeProxy {
	return &ClaudeCodeProxy{settings: settings}
}

// execution is a Claude Code run that started and ended on its own. Its texts
// already have every supplied secret hidden.
type execution struct {
	path        string
	output      string
	diagnostics string
	succeeded   bool
}

// said is everything Claude Code printed.
func (e execution) said() string {
	return strings.TrimSpace(e.output + "\n" + e.diagnostics)
}

// Inspect reports where Claude Code is and which release it is.
func (p *ClaudeCodeProxy) Inspect(ctx context.Context) (installation.InstallationVo, error) {
	run, err := p.execute(ctx, invocation{arguments: []string{"--version"}})
	if err != nil {
		return installation.InstallationVo{}, err
	}
	if !run.succeeded {
		return installation.InstallationVo{}, fmt.Errorf("%w: claude code failed: %s", installation.ErrUnrecognizedVersion, run.said())
	}
	version, err := installation.NewVersionVo(run.output)
	if err != nil {
		return installation.InstallationVo{}, err
	}
	return installation.InstallationVo{Path: run.path, Version: version}, nil
}

// ReadStatus asks Claude Code for the sign-in state. Claude Code exits with a
// failure status when nobody is signed in, so its answer, not its exit status,
// decides the outcome.
func (p *ClaudeCodeProxy) ReadStatus(ctx context.Context) (auth.StatusVo, error) {
	run, err := p.execute(ctx, invocation{arguments: []string{"auth", "status"}})
	if err != nil {
		return auth.StatusVo{}, err
	}
	payload, err := newStatusPayload([]byte(run.output))
	if err != nil {
		said := strings.TrimSpace(statusAnswer{text: run.output}.WithoutIdentity() + "\n" + run.diagnostics)
		return auth.StatusVo{}, &auth.UnreadableStatusError{Output: said}
	}
	return payload.ToVo(), nil
}

// SignIn lets the person sign in through Claude Code on the attached terminal.
func (p *ClaudeCodeProxy) SignIn(ctx context.Context, signIn auth.SignInVo) error {
	run, err := p.execute(ctx, invocation{arguments: signInCommand{signIn: signIn}.arguments(), terminal: true})
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
	run, err := p.execute(ctx, invocation{arguments: []string{"auth", "logout"}})
	if err != nil {
		return err
	}
	if !run.succeeded {
		return fmt.Errorf("%w: claude code said %q", auth.ErrSignOutFailed, run.said())
	}
	return nil
}

// Ask runs one request through Claude Code with the request on standard input,
// so requests of any length work on every platform. The final answer decides
// the outcome, whatever Claude Code's exit status.
func (p *ClaudeCodeProxy) Ask(ctx context.Context, request query.QueryVo) (query.OutcomeVo, error) {
	command := queryCommand{query: request}
	printed := &diagnostics{}
	final, answered := answer{}, false
	ask := invocation{
		arguments: command.arguments(),
		input:     request.Prompt,
		directory: request.WorkingDirectory,
		output: newAnswerStream(p.settings.Credential, printed, func(last answer) {
			final, answered = last, true
		}),
	}
	if err := command.check(); err != nil {
		return query.OutcomeVo{}, err
	}
	run, err := p.execute(ctx, ask)
	if err != nil {
		return query.OutcomeVo{}, err
	}
	if answered {
		return final.outcome, final.err
	}
	said := strings.TrimSpace(p.settings.Credential.Redact(printed.String()) + "\n" + run.said())
	if failure := (plainAnswer{text: said}).KnownFailure(); failure != nil {
		return query.OutcomeVo{}, failure
	}
	return query.OutcomeVo{}, &query.UnreadableReplyError{Output: said}
}

// Open starts a conversation: one Claude Code process that answers messages
// until it is closed. The conversation outlives ctx, which only bounds starting.
func (p *ClaudeCodeProxy) Open(ctx context.Context, request query.QueryVo) (query.IConversationLink, error) {
	if err := (queryCommand{query: request}).check(); err != nil {
		return nil, err
	}
	open := invocation{arguments: conversationCommand{query: request}.arguments(), directory: request.WorkingDirectory}
	lifetime, stop := context.WithCancel(context.WithoutCancel(ctx))
	command, err := p.command(lifetime, open)
	if err != nil {
		stop()
		return nil, err
	}
	link, err := newConversationLink(command, stop, p.settings.Credential)
	if err != nil {
		stop()
		return nil, p.startFailure(err)
	}
	return link, nil
}

// invocation is one way of running Claude Code.
type invocation struct {
	arguments []string
	// terminal attaches the run to the configured streams instead of capturing it.
	terminal bool
	// input is written to Claude Code's standard input.
	input string
	// directory is the folder Claude Code runs in; empty means the current one.
	directory string
	// output receives Claude Code's standard output as it is written instead
	// of keeping it; execution.output is then empty. Supplied secrets are not
	// hidden from it.
	output io.Writer
}

// execute runs Claude Code once. It fails only when the run could not happen or
// was given up: installation.ErrNotInstalled, the context's error, or a start
// failure. A run that ended with a failure status is reported through
// execution.succeeded. Supplied secrets are hidden from every text it returns.
func (p *ClaudeCodeProxy) execute(ctx context.Context, run invocation) (execution, error) {
	command, err := p.command(ctx, run)
	if err != nil {
		return execution{}, err
	}
	var stdout, stderr bytes.Buffer
	if run.terminal {
		command.Stdin, command.Stdout, command.Stderr = p.settings.Stdin, p.settings.Stdout, p.settings.Stderr
	} else {
		command.Stdin, command.Stdout, command.Stderr = strings.NewReader(run.input), &stdout, &stderr
	}
	if run.output != nil && !run.terminal {
		command.Stdout = run.output
	}

	runErr := command.Run()

	if ctx.Err() != nil {
		return execution{}, fmt.Errorf("claude code was given up: %w", ctx.Err())
	}
	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		return execution{}, p.startFailure(runErr)
	}
	return execution{
		path:        command.Path,
		output:      p.settings.Credential.Redact(stdout.String()),
		diagnostics: p.settings.Credential.Redact(stderr.String()),
		succeeded:   runErr == nil,
	}, nil
}

// command prepares a Claude Code invocation with the chosen executable,
// environment, work folder and cancellation.
func (p *ClaudeCodeProxy) command(ctx context.Context, run invocation) (*exec.Cmd, error) {
	path, err := p.settings.executablePath()
	if err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, path, run.arguments...)
	command.Env = p.settings.environment(os.Environ())
	command.Dir = run.directory
	command.WaitDelay = waitAfterCancel
	run.stopTogether(command)
	return command, nil
}

// startFailure translates a failure to start Claude Code, with secrets hidden.
func (p *ClaudeCodeProxy) startFailure(err error) error {
	reason := p.settings.Credential.Redact(err.Error())
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", installation.ErrNotInstalled, reason)
	}
	return errors.New("claude code could not start: " + reason)
}
