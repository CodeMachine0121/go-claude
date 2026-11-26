package claudecode

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/CodeMachine0121/go-claude/auth"
)

// statusPayload is the answer of `claude auth status`. Fields that Claude Code
// reports as null decode as empty strings.
type statusPayload struct {
	LoggedIn         *bool  `json:"loggedIn"`
	AuthMethod       string `json:"authMethod"`
	SubscriptionType string `json:"subscriptionType"`
	LoginExpiresAt   string `json:"loginExpiresAt"`
	Email            string `json:"email"`
	OrganizationName string `json:"orgName"`

	expiresAt time.Time
}

var errIncompleteStatus = errors.New("sign-in state has no signed-in flag")

// newStatusPayload decodes and validates the answer.
func newStatusPayload(data []byte) (statusPayload, error) {
	var payload statusPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return statusPayload{}, err
	}
	if payload.LoggedIn == nil {
		return statusPayload{}, errIncompleteStatus
	}
	if payload.LoginExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, payload.LoginExpiresAt)
		if err != nil {
			return statusPayload{}, err
		}
		payload.expiresAt = expiresAt
	}
	return payload, nil
}

// ToVo normalizes the answer into the auth feature's shape. Unknown sign-in
// methods are kept under the name Claude Code gave them.
func (p statusPayload) ToVo() auth.StatusVo {
	accountKind := auth.AccountKind(p.AuthMethod)
	switch p.AuthMethod {
	case "claude.ai":
		accountKind = auth.AccountKindClaudeSubscription
	case "console":
		accountKind = auth.AccountKindDeveloperAccount
	case "oauth_token":
		accountKind = auth.AccountKindSubscriptionPass
	case "api_key":
		accountKind = auth.AccountKindDeveloperAccountKey
	case "none":
		accountKind = auth.AccountKindNotSignedIn
	}
	return auth.StatusVo{
		LoggedIn:     *p.LoggedIn,
		AccountKind:  accountKind,
		Plan:         p.SubscriptionType,
		ExpiresAt:    p.expiresAt,
		Email:        p.Email,
		Organization: p.OrganizationName,
	}
}
