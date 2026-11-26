package claude_test

import (
	"context"
	"errors"
	"fmt"

	claude "github.com/CodeMachine0121/go-claude"
	"github.com/CodeMachine0121/go-claude/auth"
	"github.com/CodeMachine0121/go-claude/installation"
)

// Check the sign-in and let the person sign in with their Claude subscription.
func Example() {
	ctx := context.Background()
	client, err := claude.New(ctx)
	if errors.Is(err, installation.ErrNotInstalled) {
		fmt.Println(err)
		return
	}
	if err != nil {
		return
	}
	status, err := client.Auth().Status(ctx)
	if err == nil && !status.SignedIn {
		status, err = client.Auth().Login(ctx, auth.LoginRequest{})
	}
	if err == nil {
		fmt.Println(status.AccountKind)
	}
}

// Keep a separate sign-in for work.
func ExampleWithProfile() {
	client, err := claude.New(context.Background(), claude.WithProfile("/path/to/profiles/work"))
	if err != nil {
		return
	}
	_ = client
}

// Run unattended with a subscription pass made by `claude setup-token`.
func ExampleWithOAuthToken() {
	client, err := claude.New(context.Background(), claude.WithOAuthToken("sk-ant-oat01-..."))
	if err != nil {
		return
	}
	_ = client
}

// Supplying both kinds of credential is refused before Claude Code is touched.
func ExampleWithAPIKey_conflict() {
	_, err := claude.New(context.Background(), claude.WithAPIKey("key"), claude.WithOAuthToken("pass"))
	fmt.Println(err)
	// Output: supply only one kind of credential
}
