# Code Style

Formatting and linting follow the Go toolchain: `gofmt`, `go vet`. This file covers the
principles on top of that.

## Explicit types

- **Never use `any` / `interface{}` as a data polymorphism device.**
- When polymorphism is needed, use a named behavior interface or a sealed union
  (see [architecture.md](architecture.md)).
- Only acceptable uses: generic type parameters, and decoding loosely structured external
  output inside `internal/infrastructure` — never past the proxy boundary.

## Declare and assign at once

```go
// ❌
var total int
total = calculate()

// ✅
total := calculate()
```

Exception: decode targets that must exist empty before a decoder fills them
(`var raw statusPayload; json.Unmarshal(data, &raw)`).

## Behavior lives on objects

- Business rules are **methods on domain models**, never scattered package-level functions.
- Conversions are methods on the source (`vo.ToDomain()`, `domain.ToDto()`).
- No unexported free helper functions — apply the three-step move in
  [architecture.md](architecture.md).
- Unexported methods stay only when 2+ exported methods share them; otherwise inline.

### No helper grab-bags

**No `XxxHelper` / `XxxUtils` / `XxxManager` types or files.** Before writing one, apply the
three-step move: most helpers exist only because behavior was not placed next to its data.

Only when **all** of these hold may a helper exist:

1. Fully stateless and touches **no domain data** (if it does, it is domain behavior).
2. No existing object plausibly owns the behavior.
3. Purely technical conversion or glue (format parsing, encoding, string normalization).

Such helpers go in `internal/utilities/`, with a specific name (`ANSIStripper`, not
`CommonUtils`), and never contain business rules.

## Errors

- Errors crossing the external boundary (process failed, unreadable output) are **wrapped
  into errors the domain recognizes** — exported sentinels or named error types — so users
  can branch with `errors.Is` / `errors.As`. Never pass a raw `*exec.ExitError` up.
- Wrap with `%w` to keep the cause; add context, not noise.
- Error messages are lowercase, no trailing punctuation (Go convention).
- **Secrets never appear in errors, logs or `String()` output.** Types that carry a secret
  implement `String()` / `GoString()` to redact it.

## Context and cancellation

- Every operation that may block takes `context.Context` first and returns promptly when
  it is cancelled, reporting `ctx.Err()` (wrapped) so `errors.Is(err, context.Canceled)`
  and `errors.Is(err, context.DeadlineExceeded)` work.
- Child processes are started with `exec.CommandContext` so cancellation kills them.
