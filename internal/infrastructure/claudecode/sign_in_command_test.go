package claudecode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/CodeMachine0121/go-claude/auth"
)

func TestSignInCommand_Arguments(t *testing.T) {
	tests := []struct {
		name     string
		signIn   auth.SignInVo
		expected []string
	}{
		{
			name:     "claude subscription",
			signIn:   auth.SignInVo{Method: auth.ClaudeSubscription},
			expected: []string{"auth", "login", "--claudeai"},
		},
		{
			name:     "developer account with a pre-filled email",
			signIn:   auth.SignInVo{Method: auth.DeveloperAccount, Email: "dev@example.com"},
			expected: []string{"auth", "login", "--console", "--email", "dev@example.com"},
		},
		{
			name:     "company single sign-on",
			signIn:   auth.SignInVo{Method: auth.SingleSignOn},
			expected: []string{"auth", "login", "--sso"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, signInCommand{signIn: tt.signIn}.arguments())
		})
	}
}
