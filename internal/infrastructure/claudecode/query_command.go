package claudecode

import (
	"strconv"
	"strings"

	"github.com/CodeMachine0121/go-claude/query"
)

// queryCommand translates a validated request into `claude -p` arguments. The
// request itself is not an argument: it goes to standard input.
type queryCommand struct {
	query query.QueryVo
}

func (c queryCommand) arguments() []string {
	return append([]string{"--print", "--output-format", "json"}, c.choices()...)
}

// choices are the flags shared by one-shot requests and conversations.
func (c queryCommand) choices() []string {
	var arguments []string
	if c.query.MaxTurns > 0 {
		arguments = append(arguments, "--max-turns", strconv.Itoa(c.query.MaxTurns))
	}
	if c.query.Model != "" {
		arguments = append(arguments, "--model", c.query.Model)
	}
	if len(c.query.AllowedTools) > 0 {
		arguments = append(arguments, "--allowedTools", strings.Join(c.query.AllowedTools, ","))
	}
	if c.query.SessionID != "" {
		arguments = append(arguments, "--resume", c.query.SessionID)
	}
	return arguments
}

// conversationCommand translates a conversation into `claude -p` arguments for
// messages and answers exchanged as JSON lines.
type conversationCommand struct {
	query query.QueryVo
}

func (c conversationCommand) arguments() []string {
	conversation := []string{"--print", "--input-format", "stream-json", "--output-format", "stream-json", "--verbose"}
	return append(conversation, queryCommand{query: c.query}.choices()...)
}
