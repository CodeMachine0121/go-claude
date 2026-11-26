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
