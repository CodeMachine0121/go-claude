package query

import (
	"errors"
	"fmt"
)

// ErrEmptyRequest reports a request with nothing but spaces or line breaks.
var ErrEmptyRequest = errors.New("the request is empty")

// ErrNegativeMaxTurns reports a negative round limit.
var ErrNegativeMaxTurns = errors.New("the round limit cannot be negative")

// ErrWorkFolderMissing reports a Request.WorkingDirectory that does not exist.
var ErrWorkFolderMissing = errors.New("the work folder does not exist")

// ErrNotSignedIn reports that the person is not signed in and no pass or key
// was handed over.
var ErrNotSignedIn = errors.New("not signed in")

// ErrConversationNotFound reports a Request.SessionID that Claude Code does not know.
var ErrConversationNotFound = errors.New("conversation not found")

// ErrInvalidCost reports a cost Claude Code stated in a form that is not a
// non-negative decimal number of dollars.
var ErrInvalidCost = errors.New("cost is not a valid amount")

// ErrRoundLimitReached reports that Claude stopped at Request.MaxTurns before
// finishing. The concrete error is a *RoundLimitError.
var ErrRoundLimitReached = errors.New("round limit reached")

// ErrClaudeReportedProblem reports a problem Claude explained, such as an
// unavailable model. The concrete error is a *ProblemError.
var ErrClaudeReportedProblem = errors.New("claude reported a problem")

// ErrUnreadableReply reports an answer from Claude Code the library cannot
// understand. The concrete error is an *UnreadableReplyError.
var ErrUnreadableReply = errors.New("could not read the reply")

// RoundLimitError carries what is known when the round limit stopped Claude.
// Continue the conversation with Request.SessionID set to SessionID.
type RoundLimitError struct {
	Turns           int
	CostUSD         string
	CostNanodollars int64
	SessionID       string
}

// Error reports the rounds used.
func (e *RoundLimitError) Error() string {
	return fmt.Sprintf("%s after %d rounds", ErrRoundLimitReached, e.Turns)
}

// Unwrap returns ErrRoundLimitReached.
func (e *RoundLimitError) Unwrap() error {
	return ErrRoundLimitReached
}

// ProblemError carries Claude's explanation, with handed-over secrets hidden.
type ProblemError struct {
	Explanation string
}

// Error includes Claude's explanation.
func (e *ProblemError) Error() string {
	return fmt.Sprintf("%s: %s", ErrClaudeReportedProblem, e.Explanation)
}

// Unwrap returns ErrClaudeReportedProblem.
func (e *ProblemError) Unwrap() error {
	return ErrClaudeReportedProblem
}

// UnreadableReplyError carries what Claude Code said, with handed-over secrets hidden.
type UnreadableReplyError struct {
	Output string
}

// Error includes what Claude Code said.
func (e *UnreadableReplyError) Error() string {
	return fmt.Sprintf("%s: claude code said %q", ErrUnreadableReply, e.Output)
}

// Unwrap returns ErrUnreadableReply.
func (e *UnreadableReplyError) Unwrap() error {
	return ErrUnreadableReply
}
