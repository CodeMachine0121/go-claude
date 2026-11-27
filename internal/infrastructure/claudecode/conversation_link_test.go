package claudecode

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CodeMachine0121/go-claude/query"
)

func TestConversationWireLines(t *testing.T) {
	t.Run("a message is a user line", func(t *testing.T) {
		line, err := json.Marshal(userMessage{Type: "user", Message: userContent{Role: "user", Content: "Reply with exactly: ONE"}})

		require.NoError(t, err)
		assert.JSONEq(t, `{"type":"user","message":{"role":"user","content":"Reply with exactly: ONE"}}`, string(line))
	})

	t.Run("an interrupt is a control request", func(t *testing.T) {
		line, err := json.Marshal(interruptRequest{Type: "control_request", RequestID: "interrupt-1", Request: interruptSubtype{Subtype: "interrupt"}})

		require.NoError(t, err)
		assert.JSONEq(t, `{"type":"control_request","request_id":"interrupt-1","request":{"subtype":"interrupt"}}`, string(line))
	})
}

func TestConversationLink_ABrokenLinkRefusesEveryMessage(t *testing.T) {
	link := &conversationLink{answers: make(chan answer, 1), ended: make(chan struct{})}
	link.broken.Store(true)

	_, err := link.Send(context.Background(), "hi")

	require.ErrorIs(t, err, query.ErrConversationEnded)
}

func TestConversationLink_AnEndedContextSendsNothing(t *testing.T) {
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
	link := &conversationLink{input: writer, answers: make(chan answer, 1), ended: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = link.Send(ctx, "hi")

	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, writer.Close())
	written, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Empty(t, written, "nothing should reach claude code")
}

func TestConversationLink_GivingUpOnAStuckWriteBreaksTheLink(t *testing.T) {
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = reader.Close() })
	stopped := make(chan struct{})
	link := &conversationLink{
		input: writer, answers: make(chan answer, 1), ended: make(chan struct{}),
		stop: func() { close(stopped); _ = writer.Close() },
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	started := time.Now()

	_, err = link.Send(ctx, strings.Repeat("x", 4<<20)) // far more than a pipe holds; nobody reads it

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(started), time.Second)
	<-stopped
	assert.True(t, link.broken.Load())
	_, later := link.Send(context.Background(), "hi")
	require.ErrorIs(t, later, query.ErrConversationEnded)
}
