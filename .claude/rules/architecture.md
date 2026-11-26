# Architecture — Go Library by Feature

go-claude is a library that other Go projects `go get` and call from their own code.
Each business capability lives in its own **public feature package**; everything that
touches an external resource (the Claude Code CLI) is hidden under `internal/`.

## Directory layout

```text
go-claude/
├── doc.go                  # godoc for the root package
├── client.go               # COMPOSITION ROOT: New(opts...) wires every feature, returns *Client
├── option.go               # functional options (WithXxx)
├── auth/                   # public feature package
│   ├── doc.go              # godoc for the feature
│   ├── domain.go           # domain models (XxxDomain): business behavior lives here
│   ├── dto.go              # public API shapes (XxxRequest in, XxxDto out)
│   ├── vo.go               # value objects handed in by the proxy (XxxVo)
│   ├── errors.go           # sentinel errors / named error types of this feature
│   ├── proxy.go            # unexported interface for the external capability
│   ├── usecase.go          # business logic (UseCase struct + methods)
│   ├── mock_test.go        # testify mocks of the proxy interface (test-only)
│   ├── usecase_test.go     # UseCase tests
│   └── example_test.go     # runnable godoc examples
├── <next feature>/
├── internal/
│   └── infrastructure/
│       └── claudecode/     # ClaudeCodeProxy: the one proxy for the Claude Code CLI
├── examples/               # complete runnable programs, one main each
├── test/                   # integration tests (build tag), hit the real CLI
├── go.mod
└── README.md
```

- No `pkg/`: the module root already is the public surface.
- No `cmd/`: this module ships no binary.

## Dependency direction

```
            ┌───────────── client.go (composition root) ─────────────┐
            │ import                                        import   │
            ▼                                                        ▼
   auth/  <feature>/  ◀────── import ────── internal/infrastructure/claudecode
   (define interfaces, VOs)                 (implements feature interfaces,
                                             returns feature VOs)
```

- **infrastructure → feature, never the reverse.** The proxy normalizes raw CLI output
  into VOs owned by the feature, so it imports the feature.
- Therefore a feature can never construct its own proxy (import cycle). **The root
  package is the composition root**: the only package that imports both the features and
  `internal/infrastructure`, and the only place that calls constructors to wire
  `Proxy → UseCase`. Nothing else creates its own dependencies.
- Features should not import each other. When a concept is shared by several features
  (e.g. "the CLI cannot be found"), give that concept its own feature package named
  after it. Never create `common/`, `shared/`, `util/` grab-bags.
- The business core (feature packages) knows nothing about processes, JSON wire formats,
  environment variables or any SDK. Those details stay in infrastructure.
- Never use the word "port" for abstractions; they are "interfaces".

## Roles

| File | Role | Rules |
| :--- | :--- | :--- |
| `client.go` (root) | Composition root + public facade | `New(opts ...Option) (*Client, error)` reads options, rejects conflicting options, wires proxies into use cases. `Client` exposes each feature through an accessor (`c.Auth()`). |
| `option.go` (root) | Configuration | Functional options `WithXxx(...) Option`. They are part of the constructor, not helpers. |
| `domain.go` | Domain model | `XxxDomain` suffix. All calculation, validation and state transitions live here. Built with `NewXxxDomain(...)`; normalization happens inside it. Never put an interface in front of it. |
| `vo.go` | Value object | `XxxVo` suffix. Immutable, no business behavior. The only shape a proxy hands into a feature. May carry shape conversions such as `ToDomain()`. |
| `dto.go` | Public API shapes | `XxxRequest` (passed in by the user) / `XxxDto` (returned to the user). Data only. |
| `errors.go` | Errors | Exported sentinels (`ErrXxx`) or named error types (`*XxxError`) that users test with `errors.Is` / `errors.As`. |
| `proxy.go` | External capability interface | **Unexported**, defined by the consuming feature, named `iXxxProxy`. |
| `usecase.go` | Business logic | Exported `UseCase` struct with methods; constructor `NewUseCase(proxy ...)`. |
| `internal/infrastructure/claudecode/` | Proxy implementation | One external resource, one proxy. `ClaudeCodeProxy` may implement the interfaces of several features. |

Call chain inside a use case:

```
UseCase → iXxxProxy returns XxxVo → vo.ToDomain() → run business behavior → domain.ToDto() → user
```

Domain models and VOs are never returned to the user; always convert to a DTO first so the
public shape can stay stable while internals evolve.

## The public API is a contract

Once released, changing the public API is a breaking change.

- **Unexported by default.** Export only what users actually need.
- Every public method that does I/O takes `context.Context` as its **first parameter**;
  cancellation and timeouts always go through it.
- **No global state**: no mutable package-level variables, no `init()` side effects.
  All state hangs off `Client` / `UseCase` instances.
- Document concurrency safety in godoc. Use cases are stateless or guard their own state.
- **Every public package has a `doc.go`**; every exported identifier has a godoc comment.
- **Minimal dependencies.** Production code uses the standard library only; testify is
  allowed in `_test.go` files only.

## Interfaces: what to abstract

Follow **"accept interfaces, return structs"**.

| Thing | Interface? | Injected how |
| :--- | :--- | :--- |
| Components touching the outside (proxies) | **Yes**, unexported, defined in the consuming feature | Composition root injects the implementation; tests inject a mock |
| Internal pure business logic | **No** | Inject the real instance |
| Types handed to users (`Client`, `UseCase`) | **No** | Return concrete structs; users who need a mock define their own interface |

### Closed unions: sealed interfaces are allowed

Go has no union types. When a piece of data is **inherently one of a fixed set of shapes**
(e.g. a streamed message is an assistant, result or system message), a sealed interface
may represent it:

```go
// Message is one of: *AssistantMessage, *ResultMessage, *SystemMessage.
type Message interface {
	isMessage() // unexported marker: no other package can add an implementation
}

func (*AssistantMessage) isMessage() {}
func (*ResultMessage) isMessage()    {}
func (*SystemMessage) isMessage()    {}
```

Allowed only when **all** of these hold:

1. The marker method is unexported and every implementation lives in the same package.
2. Each variant is a plain struct; data and `ToXxx()` conversions live on the struct.
3. The godoc lists every variant; users consume it with a type switch.

This is the **only** exception to "interfaces abstract behavior, never data".

## Behavior lives next to the data it operates on

### No static-like functions on models

Go has no `static`, but a package-level function that takes a model as its only real
input is the same smell. **Domain models, VOs and DTOs carry their own behavior as
methods.**

```
❌ func StatusDtoFromVo(vo StatusVo) StatusDto   ← conversion lives on the target, as a free function
✅ func (vo StatusVo) ToDto() StatusDto           ← conversion lives on the source, as a method
✅ NewVersionVo("2.1.0")                          ← Go's idiomatic constructor, normalization inside
```

To go from A to B, write the method on A and call it `ToB()`. The conversion reads A's
fields; putting it anywhere else is Feature Envy.

### Unexported helper functions: the three-step move

When you are about to write an unexported package-level function, do not keep it as a free
function:

1. **Look at where its parameters come from** — whoever owns the parameters owns the behavior.
2. **Replace the parameters with that owner** and move the function onto it as a method.
3. Inside, **use the owner's own fields**; the parameters disappear.

### Unexported methods must clear a bar

| Used by how many exported methods | Do |
| :--- | :--- |
| **2 or more** | It earns its keep; keep it as an unexported helper method |
| **Exactly 1** | **Inline it back** into that method |

The only reasons to extract a method are removing duplication and moving behavior to its
owner — not "this is too long". Too long is a signal to split the object.

### `internal/utilities/`

The only place a helper type may live, and only for unavoidable purely technical tools
(see [code-style.md](code-style.md) for the bar). It should normally be empty.
