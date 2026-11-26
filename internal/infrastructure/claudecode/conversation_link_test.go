package claudecode

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
