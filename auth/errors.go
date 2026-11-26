package auth

import (
	"errors"
	"fmt"
)

// ErrSignInNotCompleted reports that the person did not finish signing in.
var ErrSignInNotCompleted = errors.New("sign-in was not completed")

// ErrMultipleSignInMethods reports that a LoginRequest combined several sign-in methods.
var ErrMultipleSignInMethods = errors.New("choose one sign-in method")

// ErrUnknownSignInMethod reports that a LoginRequest named a sign-in method this
// library does not define.
var ErrUnknownSignInMethod = errors.New("unknown sign-in method")

// ErrUnreadableStatus reports that Claude Code answered the sign-in question with
// something the library cannot understand. The concrete error is an
// *UnreadableStatusError.
var ErrUnreadableStatus = errors.New("could not read sign-in state")

// ErrSignOutFailed reports that Claude Code could not sign the person out.
var ErrSignOutFailed = errors.New("sign-out failed")

// UnreadableStatusError carries what Claude Code said. Supplied secrets and the
// person's identity (email, organization) are already removed from Output. It matches ErrUnreadableStatus with errors.Is.
type UnreadableStatusError struct {
	Output string
}

// Error includes what Claude Code said.
func (e *UnreadableStatusError) Error() string {
	return fmt.Sprintf("%s: claude code said %q", ErrUnreadableStatus, e.Output)
}

// Unwrap returns ErrUnreadableStatus.
func (e *UnreadableStatusError) Unwrap() error {
	return ErrUnreadableStatus
}
