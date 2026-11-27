package claudecode

import (
	"fmt"
	"strings"

	"github.com/CodeMachine0121/go-claude/query"
)

// plainAnswer is what Claude Code printed outside its JSON answers, with
// supplied secrets already hidden. Claude Code explains some failures only
// this way, before or instead of answering.
type plainAnswer struct {
	text string
}

// KnownFailure names the failure the text explains, or nil when it explains
// none the library recognises.
func (a plainAnswer) KnownFailure() error {
	switch {
	case strings.Contains(a.text, "No conversation found"):
		return fmt.Errorf("%w: claude code said %q", query.ErrConversationNotFound, a.text)
	case strings.Contains(a.text, "Invalid MCP configuration"):
		return fmt.Errorf("%w: claude code said %q", query.ErrInvalidMCPConfiguration, a.text)
	default:
		return nil
	}
}
