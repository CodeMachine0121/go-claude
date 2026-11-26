// Command ask sends one request to Claude and prints the final reply.
//
//	go run ./examples/ask -max-turns 5 "Summarise the README in this folder"
//	go run ./examples/ask -allow Read,Bash -max-turns 10 "Find TODO comments and list them"
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	claude "github.com/CodeMachine0121/go-claude"
	"github.com/CodeMachine0121/go-claude/query"
)

func main() {
	maxTurns := flag.Int("max-turns", 0, "most rounds Claude may take (0: no limit)")
	model := flag.String("model", "", "Claude model, such as sonnet or haiku")
	allow := flag.String("allow", "", "comma-separated tools Claude may use without asking, such as Read,Bash")
	session := flag.String("continue", "", "conversation reference to continue")
	flag.Parse()
	ctx := context.Background()

	client, err := claude.New(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	request := query.Request{
		Prompt:    strings.Join(flag.Args(), " "),
		MaxTurns:  *maxTurns,
		Model:     *model,
		SessionID: *session,
	}
	if *allow != "" {
		request.AllowedTools = strings.Split(*allow, ",")
	}

	result, err := client.Query().Ask(ctx, request)
	var limit *query.RoundLimitError
	switch {
	case errors.As(err, &limit):
		fmt.Printf("Stopped after %d rounds ($%s). Continue with: -continue %s\n", limit.Turns, limit.CostUSD, limit.SessionID)
		os.Exit(2)
	case errors.Is(err, query.ErrNotSignedIn):
		fmt.Fprintln(os.Stderr, "Not signed in: run `go run ./examples/signin` first.")
		os.Exit(1)
	case err != nil:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result.Text)
	fmt.Fprintf(os.Stderr, "\n%d rounds · $%s · conversation %s\n", result.Turns, result.CostUSD, result.SessionID)
}
