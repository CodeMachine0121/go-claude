# go-claude

Use Claude from Go programs through the [Claude Code](https://www.anthropic.com/claude-code)
CLI installed on the machine — with whatever account Claude Code is signed in with,
**including a Claude subscription** (Pro, Max, Team, Enterprise).

The library never reads stored credentials and never signs in on its own: every sign-in
goes through the Claude Code CLI.

> **Status:** early development. Sign-in management, profiles and one-shot requests are
> available; multi-turn conversations are next.

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

### Send a one-shot request

Claude may work over several rounds — thinking, using the tools you allow, looking at the
results — before it gives one final reply. `MaxTurns` caps the rounds (0 means no cap).

```go
result, err := client.Query().Ask(ctx, query.Request{
	Prompt:           "Find the failing test and explain why it fails",
	MaxTurns:         10,
	Model:            "sonnet",
	WorkingDirectory: "/path/to/project",
	AllowedTools:     []string{"Read", "Bash"},
})
var limit *query.RoundLimitError
switch {
case errors.As(err, &limit):
	// Stopped early; continue with query.Request{SessionID: limit.SessionID, ...}
case err != nil:
	return err // query.ErrNotSignedIn, *query.ProblemError, ...
}
fmt.Println(result.Text)
fmt.Println(result.Turns, result.CostUSD, result.SessionID)
```

> [!WARNING]
> `AllowedTools` lets Claude act inside `WorkingDirectory` without asking — running
> commands, editing files. Allow only what the task needs.

The cost is an exact decimal string (`"0.00559738"`) plus `CostNanodollars` for summing;
it is never a floating-point number. Pass `result.SessionID` as `Request.SessionID` to
continue the conversation in a later request.

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

Runnable examples live in [`examples/signin`](examples/signin) and [`examples/ask`](examples/ask).

## Using a Claude subscription: what is allowed

Using **your own** Claude subscription through Claude Code on **your own** machine is fine.

If you build a product that lets **its end users** sign in with their Claude accounts,
Anthropic requires prior approval. Without it, your product must use a developer account
key (`WithAPIKey`) and per-use billing. Check Anthropic's current terms before shipping.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
