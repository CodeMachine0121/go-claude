// Package auth reads, starts and ends the Claude Code sign-in.
//
// The library never handles passwords or stored secrets: signing in is done by
// the person, guided by the Claude Code CLI in their own terminal and browser.
// Whatever account Claude Code is signed in with — a Claude subscription, a
// developer account, or a supplied credential — is the account the program uses.
//
// Obtain a *UseCase from claude.Client.Auth:
//
//	status, err := client.Auth().Status(ctx)
//	if err == nil && !status.SignedIn {
//		status, err = client.Auth().Login(ctx, auth.LoginRequest{})
//	}
package auth
