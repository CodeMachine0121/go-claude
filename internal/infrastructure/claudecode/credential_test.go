package claudecode

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredential_Redact(t *testing.T) {
	tests := []struct {
		name       string
		credential Credential
		text       string
		expected   string
	}{
		{
			name:       "subscription pass is hidden where it appears",
			credential: NewOAuthToken("sk-ant-oat01-SECRET"),
			text:       "bad token sk-ant-oat01-SECRET here",
			expected:   "bad token [REDACTED] here",
		},
		{
			name:       "every occurrence of a developer key is hidden",
			credential: NewAPIKey("sk-ant-api03-SECRET"),
			text:       "sk-ant-api03-SECRET rejected; retry sk-ant-api03-SECRET",
			expected:   "[REDACTED] rejected; retry [REDACTED]",
		},
		{
			name:       "no credential leaves text untouched",
			credential: Credential{},
			text:       "claude code failed",
			expected:   "claude code failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.credential.Redact(tt.text))
		})
	}
}

func TestCredential_NeverPrintsItsSecret(t *testing.T) {
	credential := NewOAuthToken("sk-ant-oat01-SECRET")

	for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
		t.Run(verb, func(t *testing.T) {
			printed := fmt.Sprintf(verb, credential)

			assert.NotContains(t, printed, "sk-ant-oat01-SECRET")
			assert.Equal(t, "[REDACTED]", printed)
		})
	}
}

func TestCredential_IsZero(t *testing.T) {
	assert.True(t, Credential{}.IsZero())
	assert.False(t, NewAPIKey("sk-ant-api03-SECRET").IsZero())
}
