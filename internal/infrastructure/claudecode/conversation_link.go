package claudecode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/CodeMachine0121/go-claude/query"
)

const (
	// settleAfterGivingUp is how long a given-up message waits for Claude Code
	// to abandon the reply before the conversation is given up as broken.
	settleAfterGivingUp = 600 * time.Millisecond
	// settleAfterClosing is how long Close waits for Claude Code to finish on
	// its own before stopping it.
	settleAfterClosing = 2 * time.Second
)

// userMessage is one message written to Claude Code.
type userMessage struct {
	Type    string      `json:"type"`
	Message userContent `json:"message"`
}

type userContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// interruptRequest asks Claude Code to abandon the reply being written.
type interruptRequest struct {
	Type      string           `json:"type"`
	RequestID string           `json:"request_id"`
	Request   interruptSubtype `json:"request"`
}

type interruptSubtype struct {
	Subtype string `json:"subtype"`
}

// lineKind is the part of every line that says what it is.
type lineKind struct {
	Type string `json:"type"`
}

// conversationLink is one Claude Code process exchanging JSON lines: messages
// in, answers out. One goroutine reads every line; final answers go to the
// waiting Send, everything else is ignored.
//
// A link that could not settle a given-up message is broken: the late answer
// must never reach the next message, so every later Send fails as ended.
type conversationLink struct {
	command    *exec.Cmd
	stop       context.CancelFunc
	credential Credential
	input      io.WriteCloser
	answers    chan answer
	ended      chan struct{}
	ending     error
	broken     atomic.Bool
	requests   atomic.Int64
	closing    sync.Once
}

func newConversationLink(command *exec.Cmd, stop context.CancelFunc, credential Credential) (*conversationLink, error) {
	input, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	printed := &diagnostics{}
	command.Stderr = printed
	if err := command.Start(); err != nil {
		return nil, err
	}
	link := &conversationLink{
		command:    command,
		stop:       stop,
		credential: credential,
		input:      input,
		answers:    make(chan answer, 1),
		ended:      make(chan struct{}),
	}
	go link.read(output, printed)
	return link, nil
}

// read forwards final answers until Claude Code stops, then records why it stopped.
func (l *conversationLink) read(output io.Reader, printed *diagnostics) {
	answers := newAnswerStream(l.credential, printed, l.deliver)
	_, _ = io.Copy(answers, output)
	answers.Flush()
	_ = l.command.Wait()
	said := l.credential.Redact(strings.TrimSpace(printed.String()))
	l.ending = query.ErrConversationEnded
	if said != "" {
		l.ending = fmt.Errorf("%w: claude code said %q", query.ErrConversationEnded, said)
	}
	if strings.Contains(said, "No conversation found") {
		l.ending = fmt.Errorf("%w: claude code said %q", query.ErrConversationNotFound, said)
	}
	if strings.Contains(said, invalidMCPConfigurationWording) {
		l.ending = fmt.Errorf("%w: claude code said %q", query.ErrInvalidMCPConfiguration, said)
	}
	l.stop()
	close(l.ended)
}

// deliver hands a final answer to the waiting Send. An answer that cannot be
// understood arrives as an unreadable reply rather than being dropped, so the
// message never waits forever.
func (l *conversationLink) deliver(delivered answer) {
	select {
	case l.answers <- delivered:
	default:
	}
}

// Send writes one message and waits for its final answer. When ctx ends first
// it interrupts the reply and returns the answer that settled it, if any,
// together with the context's error.
func (l *conversationLink) Send(ctx context.Context, message string) (query.OutcomeVo, error) {
	if err := ctx.Err(); err != nil {
		return query.OutcomeVo{}, fmt.Errorf("claude code was given up: %w", err)
	}
	if l.broken.Load() {
		return query.OutcomeVo{}, query.ErrConversationEnded
	}
	line, err := json.Marshal(userMessage{Type: "user", Message: userContent{Role: "user", Content: message}})
	if err != nil {
		return query.OutcomeVo{}, err
	}
	written := make(chan error, 1)
	go func() { written <- l.write(line) }()
	select {
	case err := <-written:
		if err != nil {
			<-l.ended
			return query.OutcomeVo{}, l.ending
		}
	case <-ctx.Done():
		l.giveUp()
		return query.OutcomeVo{}, fmt.Errorf("claude code was given up: %w", ctx.Err())
	}
	select {
	case settled := <-l.answers:
		return settled.outcome, settled.err
	case <-l.ended:
		return query.OutcomeVo{}, l.ending
	case <-ctx.Done():
	}
	_ = l.Interrupt(context.Background())
	settled := answer{}
	select {
	case settled = <-l.answers:
	case <-l.ended:
	case <-time.After(settleAfterGivingUp):
		l.giveUp()
	}
	return settled.outcome, fmt.Errorf("claude code was given up: %w", ctx.Err())
}

// giveUp marks the link broken and stops Claude Code, together with the
// commands it started.
func (l *conversationLink) giveUp() {
	l.broken.Store(true)
	l.stop()
}

// Interrupt asks Claude Code to abandon the reply being written. Claude Code
// handles lines in order and ignores an interrupt when idle, so an interrupt
// written before the next message never reaches that message. A conversation
// that has ended has nothing to interrupt.
func (l *conversationLink) Interrupt(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("claude code was given up: %w", err)
	}
	line, err := json.Marshal(interruptRequest{
		Type:      "control_request",
		RequestID: "interrupt-" + strconv.FormatInt(l.requests.Add(1), 10),
		Request:   interruptSubtype{Subtype: "interrupt"},
	})
	if err != nil {
		return err
	}
	_ = l.write(line)
	return nil
}

// Close ends Claude Code's input and waits for it to finish; when it does not
// finish in time it is stopped, together with the commands it started.
// Closing the input also releases a write that Claude Code stopped reading.
func (l *conversationLink) Close() error {
	l.closing.Do(func() {
		_ = l.input.Close()
		select {
		case <-l.ended:
		case <-time.After(settleAfterClosing):
			l.stop()
			<-l.ended
		}
	})
	return nil
}

var errLinkEnded = errors.New("conversation link ended")

// write sends one JSON line in a single write, unless Claude Code has stopped.
// Lines are written whole by one call, so concurrent writes never interleave.
func (l *conversationLink) write(line []byte) error {
	select {
	case <-l.ended:
		return errLinkEnded
	default:
	}
	_, err := l.input.Write(append(line, '\n'))
	return err
}
