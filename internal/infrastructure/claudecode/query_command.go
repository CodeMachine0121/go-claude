package claudecode

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CodeMachine0121/go-claude/query"
)

// queryCommand translates a validated request into `claude -p` arguments. The
// request itself is not an argument: it goes to standard input. Answers are
// read as JSON lines so that what Claude Code states only along the way, such
// as when a reached usage limit resets, is seen too.
type queryCommand struct {
	query query.QueryVo
}

func (c queryCommand) arguments() []string {
	return append([]string{"--print", "--output-format", "stream-json", "--verbose"}, c.choices()...)
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
	if len(c.query.DisallowedTools) > 0 {
		arguments = append(arguments, "--disallowedTools", strings.Join(c.query.DisallowedTools, ","))
	}
	if paths := c.mcpConfigurationPaths(); len(paths) > 0 {
		arguments = append(append(arguments, "--mcp-config"), paths...)
	}
	if c.query.StrictMCPConfiguration {
		arguments = append(arguments, "--strict-mcp-config")
	}
	if c.query.AppendSystemPrompt != "" {
		// The joined form keeps a text that starts with "-" from reading as a flag.
		arguments = append(arguments, "--append-system-prompt="+c.query.AppendSystemPrompt)
	}
	if c.query.SessionID != "" {
		arguments = append(arguments, "--resume", c.query.SessionID)
	}
	return arguments
}

// mcpConfigurationPaths are the MCP configuration files as absolute paths; a
// relative one is found from the work folder, as Claude Code itself would. An
// absolute path also never starts with "-", so it cannot read as a flag.
func (c queryCommand) mcpConfigurationPaths() []string {
	var paths []string
	for _, path := range c.query.MCPConfigurationPaths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(c.query.WorkingDirectory, path)
		}
		if absolute, err := filepath.Abs(path); err == nil {
			path = absolute
		}
		paths = append(paths, path)
	}
	return paths
}

// check refuses a work folder or MCP configuration files that are not there,
// before anything runs.
func (c queryCommand) check() error {
	if directory := c.query.WorkingDirectory; directory != "" {
		if info, err := os.Stat(directory); err != nil || !info.IsDir() {
			return fmt.Errorf("%w: %q", query.ErrWorkFolderMissing, directory)
		}
	}
	for _, path := range c.mcpConfigurationPaths() {
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			return fmt.Errorf("%w: %q does not exist", query.ErrInvalidMCPConfiguration, path)
		}
	}
	return nil
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
