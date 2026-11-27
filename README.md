# go-claude

[![CI](https://github.com/CodeMachine0121/go-claude/actions/workflows/ci.yml/badge.svg)](https://github.com/CodeMachine0121/go-claude/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/CodeMachine0121/go-claude.svg)](https://pkg.go.dev/github.com/CodeMachine0121/go-claude)
[![Go Report Card](https://goreportcard.com/badge/github.com/CodeMachine0121/go-claude)](https://goreportcard.com/report/github.com/CodeMachine0121/go-claude)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Unofficial.** go-claude is a community project. It is not made, endorsed or supported
> by Anthropic. "Claude" and "Claude Code" are trademarks of Anthropic, PBC.

Use Claude from Go programs through the [Claude Code](https://www.anthropic.com/claude-code)
CLI installed on the machine — with whatever account Claude Code is signed in with,
**including a Claude subscription** (Pro, Max, Team, Enterprise).

The library never reads stored credentials and never signs in on its own: every sign-in
goes through the Claude Code CLI.

## Why go-claude

- **Your own sign-in.** Programs run on the Claude Code login already on the machine,
  subscription or developer key; nothing new to provision.
- **One-shot requests and conversations.** Ask once with a cap on Claude Code's rounds, or
  keep a conversation open, send message after message and interrupt a reply mid-way.
- **Cancellation that cleans up.** Giving up through `context.Context` stops Claude Code
  and, on Unix, the commands it started.
- **Exact costs.** Every reply reports its cost as decimal text and integer nanodollars,
  never a float.
- **No secrets leak.** Handed-over tokens and keys are scrubbed from every error and text.

> **Status:** pre-1.0. Sign-in management, profiles, one-shot requests and multi-message
> conversations are available; the API may still change before v1.

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
	return err // query.ErrNotSignedIn, *query.UsageLimitError, *query.ProblemError, ...
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

### Hold a conversation

A conversation keeps one Claude Code process open and answers message after message.
Claude remembers the conversation; each reply reports what it cost on its own.

```go
conversation, err := client.Query().Open(ctx, query.ConversationRequest{Model: "sonnet"})
if err != nil {
	return err
}
defer conversation.Close() // ends Claude Code's work, including commands Claude started

reply, err := conversation.Send(ctx, "Read main.go and summarise it")
fmt.Println(reply.Text, reply.CostUSD, conversation.TotalCostUSD())

// From another goroutine while Claude answers:
_ = conversation.Interrupt(ctx) // the waiting Send fails with query.ErrInterrupted
```

One message is answered at a time (`query.ErrStillAnswering` otherwise). Cancelling a
`Send`'s context gives up on that reply only; the conversation stays usable.

### Run agents: role, MCP servers, forbidden tools

`query.Request` and `query.ConversationRequest` take the same agent choices:

| Field | Claude Code flag | Effect |
| :--- | :--- | :--- |
| `AllowedTools []string` | `--allowedTools` | Tools Claude may use without asking |
| `DisallowedTools []string` | `--disallowedTools` | Tools Claude must never use, e.g. `"Edit"`, `"Bash(git push:*)"` |
| `MCPConfigurationPaths []string` | `--mcp-config` | MCP configuration files; relative paths are found from `WorkingDirectory`. A missing or unreadable file fails with `query.ErrInvalidMCPConfiguration` |
| `StrictMCPConfiguration bool` | `--strict-mcp-config` | Use only the MCP servers of `MCPConfigurationPaths`, ignoring the person's own |
| `AppendSystemPrompt string` | `--append-system-prompt` | Text added after Claude Code's system prompt, such as the agent's role |

```go
reviewer, err := client.Query().Open(ctx, query.ConversationRequest{
	WorkingDirectory:       "/agents/reviewer",
	AppendSystemPrompt:     "You are the code reviewer. Never change files.",
	MCPConfigurationPaths:  []string{"mcp.json"}, // /agents/reviewer/mcp.json
	StrictMCPConfiguration: true,
	DisallowedTools:        []string{"Edit", "Write"},
})
```

> [!NOTE]
> Claude Code records the system prompt when a conversation begins: continuing a
> conversation with `SessionID` keeps the `AppendSystemPrompt` it began with.

### When the usage limit is reached

A request or message refused by the subscription's usage limit (session, weekly or
per-model limit, or a spend cap) fails with `query.ErrUsageLimitReached`. The
`*query.UsageLimitError` says when the limit resets whenever Claude Code states it, so a
scheduler can pause instead of retrying; a conversation stays open.

```go
var usage *query.UsageLimitError
if errors.As(err, &usage) {
	if !usage.ResetsAt.IsZero() {
		time.Sleep(time.Until(usage.ResetsAt)) // or schedule the agent for later
	}
	// continue with query.Request{SessionID: usage.SessionID, ...}
}
```

> [!NOTE]
> Requests and conversations run with the person's own Claude Code settings —
> `CLAUDE.md` files, hooks, plugins and MCP servers apply as they do in Claude Code.
> Use `WithProfile` for a clean, separate setup, and `StrictMCPConfiguration` to keep
> the person's MCP servers out.

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

Runnable examples live in [`examples/signin`](examples/signin), [`examples/ask`](examples/ask)
and [`examples/chat`](examples/chat).

## Using a Claude subscription: what is allowed

Using **your own** Claude subscription through Claude Code on **your own** machine is fine.

If you build a product that lets **its end users** sign in with their Claude accounts,
Anthropic requires prior approval. Without it, your product must use a developer account
key (`WithAPIKey`) and per-use billing. Check Anthropic's current terms before shipping.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE). go-claude is not affiliated with Anthropic; using Claude through it is
governed by your own agreement with Anthropic.
