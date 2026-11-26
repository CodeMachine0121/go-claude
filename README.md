# go-claude

Use Claude from Go programs through the [Claude Code](https://www.anthropic.com/claude-code)
CLI installed on the machine — with whatever account Claude Code is signed in with,
**including a Claude subscription** (Pro, Max, Team, Enterprise).

The library never reads stored credentials and never signs in on its own: every sign-in
goes through the Claude Code CLI.

> **Status:** early development. Sign-in management and profiles are available; sending
> prompts and multi-turn conversations are next.

## Requirements

- Go 1.23 or later
- Claude Code 2.1.0 or later on the machine running your program:

```bash
npm install -g @anthropic-ai/claude-code
```

## Install

```bash
go get github.com/CodeMachine0121/go-claude
```

## Usage

### Check the sign-in and sign in if needed

```go
client, err := claude.New(ctx)
if err != nil {
	return err // installation.ErrNotInstalled, installation.ErrUnsupportedVersion, ...
}

status, err := client.Auth().Status(ctx)
if err != nil {
	return err
}
if !status.SignedIn {
	// Claude Code guides the person through the sign-in in their terminal and browser.
	status, err = client.Auth().Login(ctx, auth.LoginRequest{}) // Claude subscription by default
	if errors.Is(err, auth.ErrSignInNotCompleted) {
		return errors.New("please sign in to continue")
	}
}
fmt.Println(status.AccountKind, status.Plan, status.ExpiresAt)
```

`auth.LoginRequest{Method: auth.DeveloperAccount, Email: "dev@example.com"}` signs in with a
pay-per-use developer account; `auth.SingleSignOn` uses your company's single sign-on.

### Separate profiles

Each profile keeps its own sign-in and conversation history, separate from the everyday
Claude Code sign-in:

```go
work, err := claude.New(ctx, claude.WithProfile("/path/to/profiles/work"))
```

### Unattended use

Supply a long-lived subscription pass (from `claude setup-token`) or a developer account key.
Supplying both is refused with `claude.ErrConflictingCredentials`. Secrets never appear in
errors.

```go
client, err := claude.New(ctx, claude.WithOAuthToken(os.Getenv("CLAUDE_CODE_OAUTH_TOKEN")))
client, err := claude.New(ctx, claude.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")))
```

### Options

| Option | Effect |
| :--- | :--- |
| `WithCLIPath(path)` | Use this Claude Code executable; no other copy is tried |
| `WithProfile(directory)` | Isolated sign-in and history |
| `WithOAuthToken(token)` | Subscription pass for unattended use |
| `WithAPIKey(key)` | Developer account key for unattended use |
| `WithTerminal(in, out, err)` | Streams for the interactive sign-in (default: standard streams) |

Every operation takes a `context.Context`; cancelling it stops Claude Code promptly and the
error matches `context.Canceled` / `context.DeadlineExceeded`.

A runnable example lives in [`examples/signin`](examples/signin).

## Using a Claude subscription: what is allowed

Using **your own** Claude subscription through Claude Code on **your own** machine is fine.

If you build a product that lets **its end users** sign in with their Claude accounts,
Anthropic requires prior approval. Without it, your product must use a developer account
key (`WithAPIKey`) and per-use billing. Check Anthropic's current terms before shipping.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

TBD
