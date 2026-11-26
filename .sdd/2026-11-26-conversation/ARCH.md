# Conversation — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-11-26-conversation/PRD.md`
**Tech context:** Go 1.23+ library · library-by-feature · builds on `.sdd/2026-11-26-one-shot-query/ARCH.md`

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** `client.Query().Open(ctx, query.ConversationRequest)` starts one long-lived `claude -p --input-format stream-json --output-format stream-json` and returns a `*query.Conversation` whose `Send` writes one message and waits for its final answer.
- **Guiding principle:** **the infrastructure owns the wire (a `conversationLink`: one process, one reader goroutine, JSON lines in and out); the domain owns the conversation's state** (idle / answering / closed / ended, running cost). The domain only sees `Send → OutcomeVo`, `Interrupt`, `Close`.

---

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `query/` | **Modify** | `UseCase.Open`, `ConversationRequest`, `Conversation` (state machine, cost increase), `OutcomeInterrupted`, `CostVo.Since`, new errors. Same feature as one-shot requests: replies, costs and failures are shared vocabulary, so no feature imports another |
| `internal/infrastructure/claudecode/` | **Modify** | `ClaudeCodeProxy.Open` + `conversationLink`; `execute`'s command preparation extracted to `command(ctx, invocation)` so both paths share path, environment, work folder, cancellation and process-group behaviour; `resultPayload` recognises interrupted answers |
| `test/conversation_test.go`, README, `examples/chat/` | **Add** | Integration tests (live ones opt-in), docs, example |
| `auth/`, `installation/`, root `client.go` | **Not touched** | `Client.Query()` already exposes the use case |

---

## 3. New Classes / Modules

### `query/`

| Name | Kind | Responsibility | Collaborators | Satisfies |
| :--- | :--- | :--- | :--- | :--- |
| `UseCase.Open(ctx, ConversationRequest) (*Conversation, error)` | Use case | Validate (negative cap), start the link, wrap it in a `Conversation` | `iQueryProxy.Open`, `ConversationDomain`… | US-05 |
| `ConversationRequest{MaxTurns, Model, WorkingDirectory, AllowedTools, SessionID}` | Request | Public input; validated through the existing `QueryDomain` rules | — | US-05 |
| `iQueryProxy.Open(ctx, QueryVo) (iConversationLink, error)` | Interface (extended) | Capability: start a conversation | `ClaudeCodeProxy` | US-05 missing folder |
| `iConversationLink{Send(ctx, message) (OutcomeVo, error); Interrupt() error; Close() error}` | Unexported interface | Capability: one open conversation on the wire | `conversationLink` | US-01..04 |
| `Conversation` | Stateful session handle (like a use case bound to one conversation; not a data model) | Serialise sends (`ErrStillAnswering`), refuse after close (`ErrConversationClosed`) / end (`ErrConversationEnded`), interrupt only while answering, compute each reply's cost from the running total, expose `SessionID()`, `TotalCostUSD()` | `iConversationLink`, `OutcomeDomain`, `CostVo` | US-01..04 |
| `CostVo.Since(previous) CostVo` | VO method | Increase of a running total; a total that went down counts in full | — | US-03 |
| `OutcomeInterrupted` + `ErrInterrupted` | Outcome kind / error | An abandoned reply | `OutcomeDomain` | US-02 |
| `ErrStillAnswering`, `ErrConversationClosed`, `ErrConversationEnded` | Errors | Conversation state refusals | — | US-01, US-04 |

### `internal/infrastructure/claudecode/`

| Name | Kind | Responsibility | Satisfies |
| :--- | :--- | :--- | :--- |
| `ClaudeCodeProxy.Open(ctx, QueryVo)` | Proxy method | Refuse missing work folder; start the process with `conversationCommand` arguments; return a `conversationLink` | US-05 |
| `conversationCommand{query}.arguments()` | Pure builder | `--print --input-format stream-json --output-format stream-json --verbose` + the one-shot flags | US-05 |
| `conversationLink` | Wire session | Writes `user` and `control_request interrupt` lines (serialised); one reader goroutine decodes lines, forwards final answers, ignores everything else; on process end delivers `ErrConversationEnded` (or `ErrConversationNotFound` when Claude Code said so) to the waiting and all later sends; on give-up sends an interrupt and waits ≤1s for the answer, else stops the process; `Close` closes input, waits ≤2s, then stops the process group | US-01..05 |
| `userMessage` / `interruptRequest` | Wire encoders | JSON lines Claude Code accepts | US-01, US-02 |
| `resultPayload.ToVo` | Modified | `terminal_reason: aborted_streaming` ⇒ `OutcomeInterrupted` | US-02 |
| `command(ctx, invocation)` | Extracted | Shared preparation for `execute` and `Open` | — |

---

## 4. Modified Components

| Component | Change |
| :--- | :--- |
| `ClaudeCodeProxy.execute` | Uses the extracted `command(ctx, invocation)` |
| `resultPayload` | Reads `terminal_reason`; interrupted answers become `OutcomeInterrupted` |
| `OutcomeDomain.ToDto` | `OutcomeInterrupted` ⇒ `ErrInterrupted` |

---

## 5. Component Relationships

```mermaid
flowchart TD
    Dev[Developer] -->|Query().Open| UC[query.UseCase]
    UC -->|iQueryProxy.Open| Proxy[ClaudeCodeProxy]
    Proxy --> Link[conversationLink]
    UC --> Conv[query.Conversation]
    Dev -->|Send / Interrupt / Close| Conv
    Conv -->|iConversationLink| Link
    Conv --> OD[OutcomeDomain]
    Conv --> Cost[CostVo.Since]
    Link --> RP[resultPayload]
    Link --> CLI[(claude stream-json)]
```

---

## 6. Extensibility & Handoff Notes

- **Most likely next requirement:** seeing partial text while Claude writes, or answering Claude's permission questions during a reply.
- **Where it lands:** the reader goroutine in `conversationLink` already sees every line; partial text and permission requests become new message kinds it forwards through new link methods (e.g. a stream of events), with `Conversation` exposing them — add-only.
- **Do not hardcode:** wire field names stay in infrastructure; timeouts are named constants.
- **Known debt:** the person's own Claude Code settings apply; isolating them (setting sources) is a later option.

---

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| US-01 Claude remembers earlier messages | `conversationLink` (one process) + `Conversation.Send` |
| US-01 An empty message is refused | `Conversation.Send` (`QueryDomain` rule) → `ErrEmptyRequest`; link not called |
| US-01 A second message while Claude is still answering is refused | `Conversation` state → `ErrStillAnswering` |
| US-02 Interrupting a long reply | `Conversation.Interrupt` → `conversationLink.Interrupt` → `resultPayload` (aborted) → `ErrInterrupted` |
| US-02 Interrupting when Claude is idle does nothing | `Conversation.Interrupt` (idle ⇒ nil, link not called) |
| US-02 Giving up on one message keeps the conversation | `conversationLink.Send` (ctx done ⇒ interrupt, wait ≤1s) → wrapped `ctx.Err()` |
| US-03 A later reply costs the increase in the total | `CostVo.Since` in `Conversation.Send` |
| US-03 The first reply costs the whole total | `CostVo.Since` with zero previous |
| US-04 Closing ends Claude Code's work | `Conversation.Close` → `conversationLink.Close` (idempotent) |
| US-04 Sending after closing is refused | `Conversation` state → `ErrConversationClosed` |
| US-04 Claude Code stopping on its own ends the conversation | `conversationLink` reader (process end) → `ErrConversationEnded`; `Conversation` stays ended |
| US-05 Continuing an earlier conversation | `conversationCommand` (`--resume`) |
| US-05 An unknown earlier conversation | `conversationLink` (process end with "No conversation found") → `ErrConversationNotFound` |
| US-05 A negative round cap is refused | `UseCase.Open` (`QueryDomain`) → `ErrNegativeMaxTurns` |
| US-05 A work folder that does not exist is refused | `ClaudeCodeProxy.Open` → `ErrWorkFolderMissing` |
| US-05 A person who is not signed in | `resultPayload` ("Not logged in") → `ErrNotSignedIn`; conversation stays idle |

---

## 8. Risks & Open Decisions

- The reader must never block on a full channel: final answers go to a buffered channel of size 1 read only by the waiting `Send`; answers that arrive with nobody waiting (after a give-up) are dropped.
- Open decisions: none.
