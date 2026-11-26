package installation

import (
	"errors"
	"fmt"
)

// ErrNotInstalled reports that the Claude Code CLI could not be found, either on
// the search path or at the location chosen by the developer.
var ErrNotInstalled = errors.New(`claude code is not installed: install it with "npm install -g @anthropic-ai/claude-code"`)

// ErrUnsupportedVersion reports that the installed Claude Code CLI is older than
// MinimumSupportedVersion. The concrete error is an *UnsupportedVersionError.
var ErrUnsupportedVersion = errors.New("claude code is too old")

// ErrUnrecognizedVersion reports that the release printed by the Claude Code CLI
// could not be understood.
var ErrUnrecognizedVersion = errors.New("claude code release is not recognized")

// UnsupportedVersionError carries the release that was found and the release that
// is required. It matches ErrUnsupportedVersion with errors.Is.
type UnsupportedVersionError struct {
	Found    string
	Required string
}

// Error describes both releases.
func (e *UnsupportedVersionError) Error() string {
	return fmt.Sprintf("%s: found %s, requires %s or later", ErrUnsupportedVersion, e.Found, e.Required)
}

// Unwrap returns ErrUnsupportedVersion.
func (e *UnsupportedVersionError) Unwrap() error {
	return ErrUnsupportedVersion
}
