# Contributing

## Workflow

- Development happens on `main`.
- Commit once per completed slice of work (one behavior, one integration) — not one giant commit.
- Run before every commit:

```bash
gofmt -l . && go vet ./... && go test ./...
```

- Integration tests need a local Claude Code CLI and are opt-in:

```bash
go test -tags integration ./test/...
```

- Integration tests that send requests to Claude use your Claude usage, so they also need
  an explicit opt-in. They use the fastest model with tiny requests:

```bash
GO_CLAUDE_LIVE_TESTS=1 go test -tags integration ./test/...
```

- Tests never sign out of your everyday sign-in: sign-outs only run against fresh,
  isolated profiles after proving the profile is signed out.

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/), in English:

```
<type>(<optional scope>): <imperative summary>
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `build`, `ci`.

- Describe the change in plain words. Do not reference internal planning artifacts
  (spec documents, requirement or test-case identifiers).
- Keep the summary under 72 characters; explain the *why* in the body when it is not obvious.

## Coding rules

See [CLAUDE.md](CLAUDE.md) and [.claude/rules/](.claude/rules/).
