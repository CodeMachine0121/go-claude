// Package claude lets Go programs use Claude through the Claude Code CLI that
// is installed on the person's computer, with whatever account Claude Code is
// signed in with — including a Claude subscription.
//
// New verifies the installation and wires every feature:
//
//	client, err := claude.New(ctx, claude.WithProfile("/path/to/profile"))
//	if err != nil { ... }
//	status, err := client.Auth().Status(ctx)
//
// The library never reads stored credentials and never signs in on its own;
// every sign-in goes through Claude Code.
//
// A person using their own Claude subscription on their own computer is fine. A
// product that offers Claude account sign-in to its own end users needs prior
// approval from Anthropic; otherwise use WithAPIKey.
package claude
