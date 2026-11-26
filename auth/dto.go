package auth

import "time"

// AccountKind is how the person is signed in to Claude Code.
//
// Kinds the library does not recognise are reported with the name Claude Code
// gave them, so compare against the constants and treat anything else as
// "signed in some other way".
type AccountKind string

const (
	// AccountKindNotSignedIn means no account is signed in.
	AccountKindNotSignedIn AccountKind = "not_signed_in"
	// AccountKindClaudeSubscription means an interactive sign-in with a Claude subscription.
	AccountKindClaudeSubscription AccountKind = "claude_subscription"
	// AccountKindDeveloperAccount means an interactive sign-in with a pay-per-use developer account.
	AccountKindDeveloperAccount AccountKind = "developer_account"
	// AccountKindSubscriptionPass means a supplied long-lived subscription pass (OAuth token).
	AccountKindSubscriptionPass AccountKind = "subscription_pass"
	// AccountKindDeveloperAccountKey means a supplied developer account key (API key).
	AccountKindDeveloperAccountKey AccountKind = "developer_account_key"
)

// StatusDto is the sign-in state reported to the developer.
type StatusDto struct {
	// SignedIn is false when no account is signed in; that is an ordinary state, not a failure.
	SignedIn bool
	// AccountKind tells how the person is signed in.
	AccountKind AccountKind
	// Plan is the subscription tier (such as "pro", "max" or "team"), empty when unknown.
	Plan string
	// ExpiresAt is when the sign-in stops being valid; the zero time when it does not expire.
	ExpiresAt time.Time
	// Email of the signed-in account, empty when unknown. Never included in errors.
	Email string
	// Organization the account belongs to, empty when unknown. Never included in errors.
	Organization string
}

// SignInMethod chooses how an interactive sign-in is done. Choose exactly one;
// the zero value means ClaudeSubscription.
type SignInMethod uint8

const (
	// ClaudeSubscription signs in with a Claude subscription (Pro, Max, Team, Enterprise).
	ClaudeSubscription SignInMethod = 1 << iota
	// DeveloperAccount signs in with a pay-per-use developer account.
	DeveloperAccount
	// SingleSignOn signs in through the company's single sign-on.
	SingleSignOn
)

// LoginRequest describes an interactive sign-in.
type LoginRequest struct {
	// Method is the sign-in method; zero means ClaudeSubscription.
	Method SignInMethod
	// Email pre-fills the email address on the sign-in page; optional.
	Email string
}
