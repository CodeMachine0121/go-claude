package query

// Request is a one-shot request to Claude.
type Request struct {
	// Prompt is the request in plain words. Required.
	Prompt string
	// MaxTurns caps the rounds Claude may take. 0 means no cap.
	MaxTurns int
	// Model chooses the Claude model, such as "sonnet", "opus" or "haiku";
	// empty means Claude Code's default.
	Model string
	// WorkingDirectory is the project folder Claude works in; it must exist.
	// Empty means the program's current folder.
	WorkingDirectory string
	// AllowedTools lists the tools Claude may use without asking, such as
	// "Read", "Edit" or "Bash". Any other tool would need the person's
	// permission, which nobody can give during a one-shot request, so Claude
	// is refused when it tries one.
	AllowedTools []string
	// SessionID continues an earlier conversation; empty starts a new one.
	SessionID string
	// MCPConfigurationPaths lists files describing the MCP servers (outside
	// services) Claude may connect to. A relative path is found from
	// WorkingDirectory. Every file must exist; one Claude Code cannot read fails
	// with ErrInvalidMCPConfiguration.
	MCPConfigurationPaths []string
	// StrictMCPConfiguration makes Claude use only the MCP servers of
	// MCPConfigurationPaths, ignoring every other MCP configuration the person has.
	StrictMCPConfiguration bool
	// DisallowedTools lists the tools Claude must never use, such as "Edit" or
	// "Bash(git push:*)", even when they would otherwise be allowed.
	DisallowedTools []string
	// AppendSystemPrompt is added after Claude Code's own system prompt, such as
	// the role of an agent. Claude Code records the system prompt when a
	// conversation begins: a request continuing a conversation (SessionID) keeps
	// the text the conversation began with.
	AppendSystemPrompt string
}

// ResultDto is Claude's final reply.
type ResultDto struct {
	// Text is the final reply.
	Text string
	// Turns is the number of rounds Claude took.
	Turns int
	// SessionID identifies the conversation; pass it in Request.SessionID to continue.
	SessionID string
	// CostUSD is the cost Claude Code stated, in US dollars, as an exact decimal
	// such as "0.00559738".
	CostUSD string
	// CostNanodollars is the same cost in billionths of a dollar, for summing.
	CostNanodollars int64
}
