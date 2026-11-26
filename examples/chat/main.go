// Command chat keeps a conversation with Claude open in the terminal.
// Press Ctrl+C while Claude answers to interrupt it; type /quit to leave.
//
//	go run ./examples/chat -model sonnet
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	claude "github.com/CodeMachine0121/go-claude"
	"github.com/CodeMachine0121/go-claude/query"
)

func main() {
	model := flag.String("model", "", "Claude model, such as sonnet or haiku")
	flag.Parse()
	ctx := context.Background()

	client, err := claude.New(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	conversation, err := client.Query().Open(ctx, query.ConversationRequest{Model: *model})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conversation.Close()

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		for range interrupts {
			_ = conversation.Interrupt(ctx)
		}
	}()

	lines := bufio.NewScanner(os.Stdin)
	for fmt.Print("> "); lines.Scan(); fmt.Print("> ") {
		message := strings.TrimSpace(lines.Text())
		if message == "/quit" {
			break
		}
		if message == "" {
			continue
		}
		reply, err := conversation.Send(ctx, message)
		switch {
		case errors.Is(err, query.ErrInterrupted):
			fmt.Println("(interrupted)")
		case err != nil:
			fmt.Println("error:", err)
		default:
			fmt.Printf("%s\n  · %d rounds · $%s (total $%s)\n", reply.Text, reply.Turns, reply.CostUSD, conversation.TotalCostUSD())
		}
	}
}
