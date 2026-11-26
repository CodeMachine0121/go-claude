package claudecode

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CodeMachine0121/go-claude/auth"
)

// Recorded from Claude Code 2.1.332 (`claude auth status`); account identity replaced.
const (
	recordedSubscription = `{
  "loggedIn": true,
  "authMethod": "claude.ai",
  "apiProvider": "firstParty",
  "analyticsDisabled": false,
  "projectsDirectory": "/Users/person/.claude/projects",
  "configDirectory": "/Users/person/.claude",
  "loginExpiresAt": "2026-12-24T21:37:16.360Z",
  "consoleProfileSignInOffered": true,
  "email": "person@example.com",
  "orgId": "00000000-0000-0000-0000-000000000000",
  "orgName": "Example",
  "subscriptionType": "team",
  "storedClaudeAiLogin": true,
  "planTier": "other"
}`
	recordedNotSignedIn = `{
  "loggedIn": false,
  "authMethod": "none",
  "apiProvider": "firstParty",
  "analyticsDisabled": false,
  "projectsDirectory": "/tmp/profile/projects",
  "configDirectory": "/tmp/profile",
  "loginExpiresAt": null,
  "consoleProfileSignInOffered": true
}`
	recordedAPIKey = `{
  "loggedIn": true,
  "authMethod": "api_key",
  "apiProvider": "firstParty",
  "loginExpiresAt": null,
  "apiKeySource": "ANTHROPIC_API_KEY"
}`
	recordedOAuthToken = `{
  "loggedIn": true,
  "authMethod": "oauth_token",
  "apiProvider": "firstParty",
  "loginExpiresAt": null
}`
)

func TestNewStatusPayload_ToVo(t *testing.T) {
	tests := []struct {
		name     string
		answer   string
		expected auth.StatusVo
	}{
		{
			name:   "claude subscription on the team plan",
			answer: recordedSubscription,
			expected: auth.StatusVo{
				LoggedIn: true, AccountKind: auth.AccountKindClaudeSubscription, Plan: "team",
				ExpiresAt: time.Date(2026, time.December, 24, 21, 37, 16, 360_000_000, time.UTC),
				Email:     "person@example.com", Organization: "Example",
			},
		},
		{
			name:     "not signed in",
			answer:   recordedNotSignedIn,
			expected: auth.StatusVo{LoggedIn: false, AccountKind: auth.AccountKindNotSignedIn},
		},
		{
			name:     "developer account key",
			answer:   recordedAPIKey,
			expected: auth.StatusVo{LoggedIn: true, AccountKind: auth.AccountKindDeveloperAccountKey},
		},
		{
			name:     "subscription pass",
			answer:   recordedOAuthToken,
			expected: auth.StatusVo{LoggedIn: true, AccountKind: auth.AccountKindSubscriptionPass},
		},
		{
			name:     "developer account",
			answer:   `{"loggedIn": true, "authMethod": "console"}`,
			expected: auth.StatusVo{LoggedIn: true, AccountKind: auth.AccountKindDeveloperAccount},
		},
		{
			name:     "unfamiliar sign-in method is kept by name",
			answer:   `{"loggedIn": true, "authMethod": "workload_identity"}`,
			expected: auth.StatusVo{LoggedIn: true, AccountKind: auth.AccountKind("workload_identity")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := newStatusPayload([]byte(tt.answer))

			require.NoError(t, err)
			assert.Equal(t, tt.expected, payload.ToVo())
		})
	}
}

func TestNewStatusPayload_Unreadable(t *testing.T) {
	tests := []struct {
		name   string
		answer string
	}{
		{name: "not structured at all", answer: "garbage"},
		{name: "missing the signed-in flag", answer: `{"authMethod": "claude.ai"}`},
		{name: "expiry that is not a time", answer: `{"loggedIn": true, "loginExpiresAt": "next week"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newStatusPayload([]byte(tt.answer))

			assert.Error(t, err)
		})
	}
}
