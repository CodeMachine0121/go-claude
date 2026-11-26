package claudecode

import (
	"bufio"
	"bytes"
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
	// to abandon the reply before the conversation is stopped.
	settleAfterGivingUp = 750 * time.Millisecond
	// settleAfterClosing is how long Close waits for Claude Code to finish on
	// its own before stopping it.
	settleAfterClosing = 2 * time.Second
	// keptDiagnostics bounds the non-JSON text kept to explain an ending.
	keptDiagnostics = 4096
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

// conversationLink is one Claude Code process exchanging JSON lines: messages
// in, answers out. One goroutine reads every line; final answers go to the
// waiting Send, everything else is ignored.
type conversationLink struct {
	command    *exec.Cmd
	stop       context.CancelFunc
	credential Credential
	input      io.WriteCloser
	writing    sync.Mutex
	answers    chan resultPayload
	ended      chan struct{}
	ending     error
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
	var diagnostics bytes.Buffer
	command.Stderr = &diagnostics
	if err := command.Start(); err != nil {
		return nil, err
	}
	link := &conversationLink{
		command:    command,
		stop:       stop,
		credential: credential,
		input:      input,
		answers:    make(chan resultPayload, 1),
		ended:      make(chan struct{}),
	}
	go link.read(output, &diagnostics)
	return link, nil
}

// read forwards final answers until Claude Code stops, then records why it stopped.
func (l *conversationLink) read(output io.Reader, diagnostics *bytes.Buffer) {
	var unexpected strings.Builder
	lines := bufio.NewReader(output)
	for {
		line, err := lines.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			payload, decodeErr := newResultPayload(line)
			switch {
			case decodeErr == nil:
				select {
				case l.answers <- payload:
				default:
				}
			case !json.Valid(line) && unexpected.Len() < keptDiagnostics:
				unexpected.Write(line)
			}
		}
		if err != nil {
			break
		}
	}
	_ = l.command.Wait()
	said := l.credential.Redact(strings.TrimSpace(unexpected.String() + "\n" + diagnostics.String()))
	l.ending = query.ErrConversationEnded
	if said != "" {
		l.ending = fmt.Errorf("%w: claude code said %q", query.ErrConversationEnded, said)
	}
	if strings.Contains(said, "No conversation found") {
		l.ending = fmt.Errorf("%w: claude code said %q", query.ErrConversationNotFound, said)
	}
	l.stop()
	close(l.ended)
}

// Send writes one message and waits for its final answer.
func (l *conversationLink) Send(ctx context.Context, message string) (query.OutcomeVo, error) {
	line, err := json.Marshal(userMessage{Type: "user", Message: userContent{Role: "user", Content: message}})
	if err != nil {
		return query.OutcomeVo{}, err
	}
	if err := l.write(line); err != nil {
		<-l.ended
		return query.OutcomeVo{}, l.ending
	}
	select {
	case answer := <-l.answers:
		return answer.ToVo(l.credential), nil
	case <-l.ended:
		return query.OutcomeVo{}, l.ending
	case <-ctx.Done():
	}
	_ = l.Interrupt(context.Background())
	select {
	case <-l.answers:
	case <-l.ended:
	case <-time.After(settleAfterGivingUp):
		l.stop()
	}
	return query.OutcomeVo{}, fmt.Errorf("claude code was given up: %w", ctx.Err())
}

// Interrupt asks Claude Code to abandon the reply being written. A conversation
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
	if err := l.write(line); err != nil && !errors.Is(err, errLinkEnded) {
		return err
	}
	return nil
}

// Close ends Claude Code's input and waits for it to finish; when it does not
// finish in time it is stopped, together with the commands it started.
func (l *conversationLink) Close() error {
	l.closing.Do(func() {
		l.writing.Lock()
		_ = l.input.Close()
		l.writing.Unlock()
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

// write sends one JSON line, unless Claude Code has already stopped.
func (l *conversationLink) write(line []byte) error {
	l.writing.Lock()
	defer l.writing.Unlock()
	select {
	case <-l.ended:
		return errLinkEnded
	default:
	}
	_, err := l.input.Write(append(line, '\n'))
	return err
}
