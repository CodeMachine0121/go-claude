# Agent-Team Request Options — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-11-27-agent-team-request-options/PRD.md`
**Tech context:** Go 1.23+ library · library-by-feature · builds on `.sdd/2026-11-26-one-shot-query/ARCH.md` and `.sdd/2026-11-26-conversation/ARCH.md`

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** `query.Request` and `query.ConversationRequest` gain four choices that `queryCommand` turns into Claude Code flags, and both one-shot requests and conversations read Claude Code's answer **stream** through one `answerStream`, which recognises a usage-limit refusal and the reset moment Claude Code states.
- **Guiding principle:** **the wire knows the reset moment, the domain knows what it means.** Infrastructure normalises "refused by the usage limit, resets at T" into `OutcomeVo{Kind: OutcomeUsageLimitReached, ResetsAt: T}`; `OutcomeDomain` turns it into `*UsageLimitError`. One reader of Claude Code's answer lines serves both paths, so a new line kind is learnt once.

---

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `query/dto.go`, `query/conversation.go` | **Modify** | `Request` and `ConversationRequest` gain `MCPConfigurationPaths`, `StrictMCPConfiguration`, `DisallowedTools`, `AppendSystemPrompt` (US-02..05) |
| `query/vo.go`, `query/domain.go` | **Modify** | `QueryVo` carries the four choices (slices copied by the domains); `OutcomeUsageLimitReached`, `OutcomeVo.ResetsAt`; `OutcomeDomain.ToDto` maps the new kind (US-01) |
| `query/errors.go` | **Modify** | `ErrUsageLimitReached` + `*UsageLimitError`; `ErrInvalidMCPConfiguration` |
| `internal/infrastructure/claudecode/query_command.go` | **Modify** | Flags for the four choices; outside-service lists resolved against the work folder; `checkMCPConfigurations` refuses missing lists before starting |
| `internal/infrastructure/claudecode/answer_stream.go` | **Add** | `answerStream`: an `io.Writer` that splits Claude Code's output into lines, keeps non-JSON lines as bounded diagnostics, follows usage-limit notices and hands each final answer to a callback |
| `internal/infrastructure/claudecode/result_payload.go` | **Modify** | Reads `api_error`; `usage_limit_reached` (and the older "Claude AI usage limit reached\|epoch" wording) ⇒ `OutcomeUsageLimitReached` with the stated reset |
| `internal/infrastructure/claudecode/claude_code_proxy.go` | **Modify** | `Ask` uses `--output-format stream-json --verbose` and an `answerStream` as standard output (nothing but the last answer is kept); `execute` accepts a caller-supplied output writer; missing / invalid lists ⇒ `ErrInvalidMCPConfiguration` |
| `internal/infrastructure/claudecode/conversation_link.go` | **Modify** | Reading moves into `answerStream`; an ending that says "Invalid MCP configuration" ⇒ `ErrInvalidMCPConfiguration` |
| README, UL-MAP, `test/` | **Modify / Add** | Field docs; integration checks (missing list without usage; live flags opt-in) |
| `auth/`, `installation/`, root `client.go`, `option.go` | **Not touched** | Choices are per request, not per client |

---

## 3. New Classes / Modules

### `query/`

| Name | Kind | Responsibility | Satisfies |
| :--- | :--- | :--- | :--- |
| `Request` / `ConversationRequest` `.MCPConfigurationPaths []string`, `.StrictMCPConfiguration bool`, `.DisallowedTools []string`, `.AppendSystemPrompt string` | Request fields | Public input; zero values change nothing | US-02..05 |
| `OutcomeUsageLimitReached`, `OutcomeVo.ResetsAt time.Time` | Outcome kind / VO field | Normalised refusal; zero `ResetsAt` = not stated | US-01 |
| `UsageLimitError{ResetsAt, Explanation, Turns, CostUSD, CostNanodollars, SessionID}` | Named error, `Unwrap() → ErrUsageLimitReached` | What a scheduler needs: when to resume and which conversation | US-01 |
| `ErrInvalidMCPConfiguration` | Sentinel | Missing or unreadable outside-service list | US-02 |

### `internal/infrastructure/claudecode/`

| Name | Kind | Responsibility | Satisfies |
| :--- | :--- | :--- | :--- |
| `answerStream` (`newAnswerStream(credential, printed, deliver)`, `Write`, `Flush`) | Deep module (`io.Writer`) | Hides line splitting, partial lines, non-JSON diagnostics, usage-limit notices (`rate_limit_event` status `rejected` / `allowed`, assistant `api_error_params.rate_limit_info`) and final-answer decoding; callers only receive `answer{outcome, err}` | US-01, US-05 |
| `streamLine` / `rateLimitInfo` | Wire decoders | The parts of a line that matter: type, rate-limit info, error params | US-01 |
| `queryCommand.check()` / `mcpConfigurationPaths()` | Command methods | Absolute list paths (relative ones from the work folder); refuse a missing work folder or list before starting | US-02 |

---

## 4. Modified Components

| Component | Change |
| :--- | :--- |
| `queryCommand.choices()` | `--mcp-config <abs paths…>`, `--strict-mcp-config`, `--disallowedTools a,b`, `--append-system-prompt=<text>` (the `=` form keeps text beginning with `-` from reading as a flag) |
| `queryCommand.arguments()` | `--print --output-format stream-json --verbose` |
| `resultPayload` | `APIError` field, `resetsAt` set by the stream, new kind in `ToVo` |
| `ClaudeCodeProxy.execute` / `invocation` | `invocation.output io.Writer` replaces the stdout buffer when set |
| `conversationLink.read` | `io.Copy(answerStream, stdout)`; the `diagnostics` type is shared with the stream |

---

## 5. Data Flow

```mermaid
sequenceDiagram
  participant U as UseCase / Conversation
  participant P as ClaudeCodeProxy
  participant S as answerStream
  participant C as claude -p (stream-json)
  U->>P: Ask / Open(QueryVo)
  P->>P: queryCommand.check (work folder, MCP files)
  P->>C: flags + prompt
  C-->>S: lines (assistant, rate_limit_event, result, plain text)
  S->>S: remember rejected reset; keep plain text
  S-->>P: answer{OutcomeVo{Kind, ResetsAt}}
  P-->>U: OutcomeVo
  U->>U: OutcomeDomain.ToDto → *UsageLimitError
```

---

## 6. Axis of Change & Extensibility

- **Next likely requirement:** another typed refusal Claude Code already names (e.g. `model_requires_usage_credits` ⇒ "switch model"), or exposing usage warnings before the limit. Seam: `answerStream` already sees every line and `resultPayload.ToVo` already switches on `api_error`; a new kind is one `case` plus one `OutcomeKind` and one `ToDto` branch.
- **Next likely flag:** one field on both requests + `QueryVo` + one line in `queryCommand.choices()`. The two request types stay flat (no embedded struct) to keep the public shape consistent with existing fields.
- **Rejected alternative:** parsing the reset moment from the refusal text ("resets 3pm (Europe/London)") — local, wording-dependent, no date. The stream carries unix seconds.

---

## 7. Traceability

| PRD scenario | Components |
| :--- | :--- |
| US-01 A one-shot request refused by the usage limit reports the reset moment | `answerStream`, `resultPayload.ToVo`, `OutcomeDomain.ToDto`, `UsageLimitError` |
| US-01 No reset moment is invented when Claude Code states none | `answerStream` (no rejected notice ⇒ zero), `UsageLimitError.Error` |
| US-01 A conversation message refused by the usage limit leaves the conversation open | `conversationLink` + `answerStream`, `Conversation.Send` (idle after ordinary outcomes) |
| US-01 Another problem is never mistaken for the usage limit | `resultPayload.ToVo` (keyed on `api_error`, not on text or 429) |
| US-01 Usage close to the limit is not a failure | `answerStream` ignores `allowed` / `allowed_warning`; `resultPayload.ToVo` |
| US-02 Only the listed outside services are used | `queryCommand.choices` (`--mcp-config`, `--strict-mcp-config`) |
| US-02 Listed services join the person's own | `queryCommand.choices` (no strict flag) |
| US-02 A list that cannot be found is refused before starting | `queryCommand.check`, `ClaudeCodeProxy.Ask/Open` |
| US-02 A list that cannot be read fails | `plainAnswer.KnownFailure` used by `ClaudeCodeProxy.Ask` and `conversationLink.read` ("Invalid MCP configuration") |
| US-02 A list named by its place in the project is found in the work folder | `queryCommand.mcpConfigurationPaths` |
| US-03 Forbidden actions are unavailable | `queryCommand.choices` (`--disallowedTools`) |
| US-03 Nothing forbidden keeps every action | `queryCommand.choices` (flag omitted) |
| US-04 Role instructions are followed | `queryCommand.choices` (`--append-system-prompt`) |
| US-04 No role instructions keeps the usual instructions | `queryCommand.choices` (flag omitted) |
| US-04 A continued conversation keeps the role it began with | Claude Code behaviour; godoc on `AppendSystemPrompt` |
| US-05 A conversation carries every choice | `ConversationDomain`, `conversationCommand.arguments` |

---

## 7a. Test Plan

- `query/`: `UseCase.Ask` / `Conversation.Send` with mocked proxy returning `OutcomeUsageLimitReached` (with and without reset); request fields reach the proxy.
- `claudecode/`: `answerStream` fed recorded lines (usage-limit refusal, rejected event, allowed event, split writes, plain text); `resultPayload` with recorded refusal; `queryCommand` arguments and list resolution / missing list.
- `test/`: missing list refused (no usage); invalid list fails (no usage); live, opt-in: codeword via role instructions, strict empty list, forbidden tools.
