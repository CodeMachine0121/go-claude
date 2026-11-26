// Command signin shows the sign-in state of Claude Code and offers to sign in.
//
//	go run ./examples/signin                 # everyday sign-in
//	go run ./examples/signin -profile ./work  # isolated profile
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	claude "github.com/CodeMachine0121/go-claude"
	"github.com/CodeMachine0121/go-claude/auth"
	"github.com/CodeMachine0121/go-claude/installation"
)

func main() {
	profile := flag.String("profile", "", "profile directory (empty: everyday sign-in)")
	flag.Parse()
	ctx := context.Background()

	options := []claude.Option{}
	if *profile != "" {
		options = append(options, claude.WithProfile(*profile))
	}
	client, err := claude.New(ctx, options...)
	if errors.Is(err, installation.ErrNotInstalled) || errors.Is(err, installation.ErrUnsupportedVersion) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot use claude code:", err)
		os.Exit(1)
	}
	fmt.Printf("Claude Code %s at %s\n", client.Installation().Version, client.Installation().Path)

	status, err := client.Auth().Status(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if status.SignedIn {
		fmt.Printf("Signed in (%s, plan %q)\n", status.AccountKind, status.Plan)
		return
	}

	fmt.Print("Not signed in. Sign in with your Claude subscription now? [y/N] ")
	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	if strings.TrimSpace(strings.ToLower(answer)) != "y" {
		return
	}
	status, err = client.Auth().Login(ctx, auth.LoginRequest{Method: auth.ClaudeSubscription})
	if errors.Is(err, auth.ErrSignInNotCompleted) {
		fmt.Println("Sign-in was not completed.")
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Signed in (%s, plan %q)\n", status.AccountKind, status.Plan)
}
