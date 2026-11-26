// Package query sends a one-shot request to Claude through Claude Code and
// returns the final reply.
//
// Claude may work on a request over several rounds — thinking, using the tools
// allowed in Request.AllowedTools, looking at the results — before it answers.
// Request.MaxTurns caps those rounds. Only the final reply is returned.
//
// Obtain a *UseCase from claude.Client.Query:
//
//	result, err := client.Query().Ask(ctx, query.Request{Prompt: "Summarise README.md", MaxTurns: 5})
//	var limit *query.RoundLimitError
//	if errors.As(err, &limit) {
//		// continue later with query.Request{SessionID: limit.SessionID, ...}
//	}
//
// Allowing tools lets Claude act inside Request.WorkingDirectory without asking:
// running commands, editing files. Allow only what the task needs.
package query
