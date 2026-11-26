# Naming Conventions

Casing and file-name format follow Go conventions (exported `PascalCase`, unexported
`camelCase`, files `snake_case.go`). This file decides **what** things are called.

## General

- **Full words, no abbreviations** (`SubscriptionType`, not `SubType`; `ConfigurationDirectory`,
  not `CfgDir`). Well-known initialisms keep Go casing: `CLI`, `ID`, `URL`, `JSON`, `OAuth`.
- **Do not repeat the package name** in identifiers: `auth.UseCase`, `auth.StatusDto`,
  `auth.ErrSignInNotCompleted` — never `auth.AuthUseCase`.
- **No invented synonyms.** Business vocabulary comes from `.sdd/UL-MAP.md`; a new term goes
  into the map first, then into code.

## Role suffixes

| Role | Suffix | Example |
| :--- | :--- | :--- |
| Business logic of a feature | `UseCase` | `auth.UseCase` |
| Wrapper around an external resource | `Proxy` | `claudecode.ClaudeCodeProxy` |
| Domain model (behavior lives here) | `Domain` | `auth.SignInStatusDomain` |
| Value object (immutable, no behavior) | `Vo` | `auth.StatusVo`, `auth.VersionVo` |
| Data returned to the user | `Dto` | `auth.StatusDto` |
| Data passed in by the user | `Request` | `auth.LoginRequest` |
| Functional option | `With` prefix | `claude.WithProfile` |
| Sentinel error | `Err` prefix | `auth.ErrNotInstalled` |
| Named error type | `Error` suffix | `auth.UnsupportedVersionError` |

There is no entity role: the library owns no database. Raw external formats are decoded
inside infrastructure and normalized into VOs.

### External resources always end in `Proxy`

Anything that calls an external resource — a process, an HTTP API, a third-party SDK — is
wrapped by an object ending in `Proxy`.

- **Never** use `Client`, `Gateway`, `Adapter`, `Api`, `Connector`, `Runner`, `Executor`
  for that role. (The root `claude.Client` is the library's public facade, not a wrapper
  around an external resource.)
- **One external resource, one proxy.** Every call to the Claude Code CLI goes through
  `ClaudeCodeProxy`; no other package spawns the CLI.
- The proxy normalizes raw output into VOs; wire formats never leak into features.

## Interfaces: name the capability, not the provider

- Interfaces describe **behavior contracts** only, with the single exception of sealed
  unions (see [architecture.md](architecture.md)).
- Proxy interfaces are **unexported** and use the `i` prefix: `iXxxProxy`.
- Name the **capability**, not who provides it today:

```
❌ iClaudeCodeProxy   ← names the provider
✅ iSignInProxy        ← names the capability; ClaudeCodeProxy implements it
```

Test: if the provider changed tomorrow, the interface name must not need to change.

- An interface lives in the `proxy.go` of the feature that consumes it.
- Implementation files never declare interfaces.

### Data is never an interface

Every data shape — domain model, VO, DTO, request — is a `struct`. When a use case needs a
group of parameters, wrap them in a `XxxRequest` struct; never an interface, a map, or a
long positional parameter list.

## Implementations

Implementations carry the plain role name, prefixed only by a provider name when the
provider is the meaningful distinction: `ClaudeCodeProxy`. Never a transport prefix
(`ExecProxy`, `ProcessProxy`).

## File names

- Inside a feature package, files are named by **role**: `domain.go`, `vo.go`, `dto.go`,
  `errors.go`, `proxy.go`, `usecase.go`, `doc.go`.
- When a role file grows to hold several unrelated main types, split by type:
  `status_domain.go`, `version_vo.go`.
- Infrastructure files are named after their main type: `claude_code_proxy.go`.
- Mocks live in `mock_test.go`; examples in `example_test.go`.
