# CLAUDE.md

This file only tells you **which rule in `.claude/rules/` to read in which situation.**
Product knowledge lives elsewhere: domain vocabulary in `.sdd/UL-MAP.md`, user docs in
`README.md`, workflow in `CONTRIBUTING.md`.

Everything in this repository — code, comments, docs, commit messages — is written in English.

## Read before writing code

- [.claude/rules/architecture.md](.claude/rules/architecture.md) — where code goes and which way dependencies point
- [.claude/rules/naming.md](.claude/rules/naming.md) — what things are called

## Situation table

| I am about to… | Read |
| :--- | :--- |
| Add a new capability / decide which package code belongs in | [architecture.md](.claude/rules/architecture.md) (layout, roles) |
| Call the Claude Code CLI or any external resource | [architecture.md](.claude/rules/architecture.md), [naming.md](.claude/rules/naming.md) (one `Proxy` per resource, lives in `internal/infrastructure`) |
| Export a new type, function or method | [architecture.md](.claude/rules/architecture.md) (the public API is a contract) |
| Add a domain model, VO, DTO or request | [architecture.md](.claude/rules/architecture.md), [naming.md](.claude/rules/naming.md) |
| Write a conversion between shapes | [architecture.md](.claude/rules/architecture.md) (`source.ToTarget()`, never a free function) |
| Write an unexported free function or a one-use unexported method | [architecture.md](.claude/rules/architecture.md) (three-step move / inline bar) |
| Define an interface | [naming.md](.claude/rules/naming.md) (unexported `iXxxProxy`, capability-named) |
| Model data that is one of several shapes | [architecture.md](.claude/rules/architecture.md) (sealed union rules) |
| Create a `Helper` / `Utils` type | [code-style.md](.claude/rules/code-style.md) (forbidden unless it clears the bar) |
| Return or wrap an error, handle a secret, honor cancellation | [code-style.md](.claude/rules/code-style.md) |
| Write tests or mocks | [testing.md](.claude/rules/testing.md) |
| Commit | [CONTRIBUTING.md](CONTRIBUTING.md) |

## Non-negotiables

1. **Dependencies point at the features**; infrastructure imports features, never the reverse. Only the root package wires them.
2. **Every call to the Claude Code CLI goes through `ClaudeCodeProxy`.**
3. **The library never reads stored credentials and never signs in on its own**; all sign-in goes through the CLI.
4. **Behavior lives next to its data**: methods on models, `source.ToTarget()` conversions, no helper grab-bags.
5. **Interfaces abstract behavior, not data** (sealed unions are the only exception).
6. **Interfaces are named for the capability, not the provider.**
7. **Public API is unexported by default**, takes `context.Context` first, holds no global state.
8. **Tests verify business behavior**, mock only the proxy with testify, never hand-write fakes.
