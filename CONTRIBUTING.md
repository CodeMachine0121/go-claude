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
