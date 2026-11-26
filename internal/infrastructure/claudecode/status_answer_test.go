package claudecode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusAnswer_WithoutIdentity(t *testing.T) {
	t.Run("a structured answer loses the person's identity but keeps the rest", func(t *testing.T) {
		shown := statusAnswer{text: recordedSubscription}.WithoutIdentity()

		assert.NotContains(t, shown, "person@example.com")
		assert.NotContains(t, shown, "Example")
		assert.NotContains(t, shown, "00000000-0000-0000-0000-000000000000")
		assert.Contains(t, shown, `"authMethod":"claude.ai"`)
		assert.Contains(t, shown, `"loginExpiresAt":"2026-12-24T21:37:16.360Z"`)
	})

	t.Run("text that is not structured is shown as-is", func(t *testing.T) {
		assert.Equal(t, "garbage", statusAnswer{text: "garbage"}.WithoutIdentity())
	})
}
