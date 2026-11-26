package claudecode

import "strings"

const (
	oauthTokenVariable = "CLAUDE_CODE_OAUTH_TOKEN"
	apiKeyVariable     = "ANTHROPIC_API_KEY"
	redactedMark       = "[REDACTED]"
)

// Credential is a secret handed to Claude Code: either a subscription pass
// (OAuth token) or a developer account key (API key). It never prints its secret.
type Credential struct {
	variable string
	secret   string
}

// NewOAuthToken wraps a long-lived Claude subscription pass.
func NewOAuthToken(secret string) Credential {
	return Credential{variable: oauthTokenVariable, secret: secret}
}

// NewAPIKey wraps a developer account key.
func NewAPIKey(secret string) Credential {
	return Credential{variable: apiKeyVariable, secret: secret}
}

// IsZero reports whether no credential was supplied.
func (c Credential) IsZero() bool {
	return c.secret == ""
}

// Redact hides every occurrence of the secret in text.
func (c Credential) Redact(text string) string {
	if c.IsZero() {
		return text
	}
	return strings.ReplaceAll(text, c.secret, redactedMark)
}

// String never reveals the secret.
func (c Credential) String() string {
	return redactedMark
}

// GoString never reveals the secret, including under the %#v verb.
func (c Credential) GoString() string {
	return redactedMark
}
