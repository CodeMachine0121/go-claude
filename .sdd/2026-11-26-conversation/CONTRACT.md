# Contract Traceability Matrix — 2026-11-26-conversation

Contract: PRD.md (v1.0, Finalized)
Design map: ARCH.md
Implementation: `query/` (conversation.go, usecase.go, proxy.go, vo.go, domain.go), `internal/infrastructure/claudecode/` (conversation_link.go, claude_code_proxy.go, query_command.go, result_payload.go)
Oracle: Acceptance Criteria (30 clauses: 16 AC, 9 BR, 5 NFR)
Audit date: 2026-11-26 · static audit + throwaway probes (scratchpad, `sh` scripts standing in for the CLI; Claude Code never started) · unit tests `go test -race ./query/ ./internal/...` green · integration tests not run

## Clauses

The `Spec-expected` column holds the business-observable oracle from Phase 2 (written before any code was opened). `Test audit` / `Code audit` are Claude's shadow verdict, recorded before Jev ran. `Jev test` / `Jev code`, `Status` and `Route` are copied as returned by `jev_audit.py`. Where Claude's shadow disagrees with Jev, the Notes column gives the evidence.

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev test | Jev code | Status | Route |
|----|--------|------------------------|------|------|------------|------------|----------|----------|--------|-------|
| AC-1 | Claude remembers earlier messages — sends "Reply with exactly: ONE" then "What did you just reply? One word." → first reply "ONE", second contains "ONE" | First reply text is ONE; second reply text contains ONE | query/conversation.go:87-125; conversation_link.go:131-143 | test/conversation_test.go:77 TestClaudeRemembersEarlierMessages (live); query/conversation_test.go:62 | asserts-oracle | produces-oracle | asserts-oracle 0.89 | unclear 0.60 | ❔ unclear | escalate (disagree) |
| AC-2 | An empty message is refused — only spaces → "the request is empty"; next message answered normally | Refused as empty without reaching Claude; next message answered | query/conversation.go:88-90; query/domain.go:12-14 | query/conversation_test.go:100 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.92 | ✅ conforms | accept |
| AC-3 | A second message while Claude is still answering is refused — "still answering"; first reply still arrives | Second refused as still answering; first reply delivered to its sender | query/conversation.go:91-97, 57-68 | query/conversation_test.go:115 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.40 | ✅ conforms | escalate (low confidence) |
| AC-4 | Interrupting a long reply — essay fails "interrupted"; next "Reply with exactly: AFTER" → "AFTER" | Essay message fails as interrupted; next reply is AFTER | query/conversation.go:130-138; conversation_link.go:160-176; result_payload.go:54-55; query/domain.go:59-60 | query/conversation_test.go:139; test/conversation_test.go:102 (live) | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.68 | ✅ conforms | recheck (mid confidence) |
| AC-5 | Interrupting when Claude is idle does nothing — no failure; next message answered normally | Interrupt reports no failure; next message answered | query/conversation.go:130-136 | query/conversation_test.go:165 | shallow | produces-oracle | shallow 0.80 | produces-oracle 0.93 | 🟠 mis-asserted | recheck (mid confidence) |
| AC-6 | Giving up on one message keeps the conversation — reported as given up within one second; next "AFTER" → "AFTER" | Given-up reported ≤1s after giving up; next reply AFTER | conversation_link.go:146-155; query/conversation.go:99-113 | test/conversation_test.go:122 (live); query/conversation_test.go:176 | asserts-oracle | produces-oracle | asserts-oracle 0.83 | produces-oracle 0.16 | ✅ conforms | escalate (low confidence) |
| AC-7 | A later reply costs the increase in the total — 0.0038563 so far, stated 0.00612419 → reply "0.00226789", total "0.00612419" | Reply cost 0.00226789 exactly; total 0.00612419 | query/vo.go:85-90; query/conversation.go:114-119 | query/conversation_test.go:62, 289 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.77 | ✅ conforms | recheck (mid confidence) |
| AC-8 | The first reply costs the whole total — stated 0.0038563 → reply "0.0038563", total "0.0038563" | Reply cost and total both 0.0038563 | query/vo.go:85-90; query/conversation.go:114-119 | query/conversation_test.go:62, 289 | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.56 | ✅ conforms | escalate (low confidence) |
| AC-9 | Closing ends Claude Code's work — no longer running; closing again reports no failure | Claude Code not running after Close; second Close no failure | query/conversation.go:142-151; conversation_link.go:180-193 | test/conversation_test.go:154 (live, pgrep); query/conversation_test.go:192 | asserts-oracle | produces-oracle | asserts-oracle 0.83 | produces-oracle 0.89 | ✅ conforms | recheck (mid confidence) |
| AC-10 | Sending after closing is refused — "the conversation is closed" | Refused as closed | query/conversation.go:91-95 | query/conversation_test.go:203; test/conversation_test.go:43 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.99 | ✅ conforms | accept |
| AC-11 | Claude Code stopping on its own ends the conversation — fails "the conversation ended", every later message too | This and every later message fail as ended | conversation_link.go:114-129, 137-145; query/conversation.go:106-108 | query/conversation_test.go:236 | asserts-oracle | produces-oracle | asserts-oracle 0.93 | produces-oracle 0.91 | ✅ conforms | accept |
| AC-12 | Continuing an earlier conversation — reply contains "OK" | Reply text contains OK | query_command.go:32-34, 44-47; query/conversation.go:33-39 | test/conversation_test.go:139 (live) | asserts-oracle | produces-oracle | asserts-oracle 0.95 | unclear 0.29 | ❔ unclear | escalate (disagree) |
| AC-13 | An unknown earlier conversation — fails "conversation not found" | Message fails as conversation not found | conversation_link.go:124-126; result_payload.go:56-57; query/conversation.go:106-124 | query/conversation_test.go:250, 263; test/conversation_test.go:54 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.93 | ✅ conforms | accept |
| AC-14 | A negative round cap is refused — "the round limit cannot be negative" | Opening refused, nothing started | query/conversation.go:30-32; query/usecase.go:33-36 | query/conversation_test.go:40 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.96 | ✅ conforms | accept |
| AC-15 | A work folder that does not exist is refused — "the work folder does not exist" | Opening refused before starting | claude_code_proxy.go:133-136, 163-171 | query/conversation_test.go:51; test/conversation_test.go:34 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.96 | ✅ conforms | accept |
| AC-16 | A person who is not signed in — fails "not signed in"; conversation stays open | Message fails as not signed in; next message still reaches Claude Code | result_payload.go:58-59; query/conversation.go:110-124 | query/conversation_test.go:276; test/conversation_test.go:65 (live) | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.91 | ✅ conforms | accept |
| BR-1 | One message is answered at a time. | Never two messages in flight; extra refused | query/conversation.go:91-97 | query/conversation_test.go:115 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | unclear 0.35 | ❔ unclear | escalate (disagree) |
| BR-2 | The round cap applies to each message; 0 means none; negative is refused when opening. | Each message gets its own cap and reports round limit when hit; 0 = no cap; negative refused at open | query_command.go:23-25, 44-47; result_payload.go:52-53; query/conversation.go:30-32 | query/conversation_test.go:85, 40; query_command_test.go:55-77 | shallow | unclear | asserts-oracle 0.81 | produces-oracle 0.56 | ✅ conforms | escalate (disagree) |
| BR-3 | A reply's cost is the increase of the conversation's total; the first reply costs the whole total; a total that went down makes the reply cost the new total. | Every reply moves the running total to Claude Code's stated total; each reply's cost = its own increase | query/vo.go:85-90; query/conversation.go:110-119; conversation_link.go:149-150 | query/conversation_test.go:289, 62, 85 | shallow | diverges | asserts-oracle 0.48 | produces-oracle 0.48 | ✅ conforms | escalate (disagree) |
| BR-4 | Interrupting affects only the reply being written; when idle it does nothing. | Interrupt aborts only the message in flight at the time of the call; idle → nothing | query/conversation.go:131-137 | query/conversation_test.go:138-173 | asserts-oracle | diverges | asserts-oracle 0.93 | produces-oracle 0.84 | ✅ conforms | escalate (disagree) |
| BR-5 | Giving up interrupts and reports given up; if Claude Code does not settle within one second, the conversation ends. | Given up ≤1s; when unsettled, the conversation ends: later messages fail as ended, never get a stale answer | conversation_link.go:106-109, 146-155; query/conversation.go:106-113 | test/conversation_test.go:122 (live, settled path only); query/conversation_test.go:176 | shallow | diverges | asserts-oracle 0.21 | produces-oracle 0.28 | ✅ conforms | escalate (disagree) |
| BR-6 | Closing ends Claude Code's work, including commands Claude started; it is harmless to repeat. | After Close (bounded time) neither Claude Code nor commands it started run; repeat is harmless | conversation_link.go:180-193, 198-207; process_group_unix.go:19-27 | test/conversation_test.go:154 (pgrep matches Claude Code only); query/conversation_test.go:192 | shallow | unclear | asserts-oracle 0.95 | produces-oracle 0.71 | ✅ conforms | escalate (disagree) |
| BR-7 | Every failure that a one-shot request recognises (empty, round limit, not signed in, problem, unknown conversation, missing work folder) is recognised per message the same way. | Each of the six failures surfaces per message exactly as for one-shot | query/domain.go:40-66; query/conversation.go:110-124; result_payload.go:48-71 | query/conversation_test.go:85-113, 250-287, 51 | shallow | produces-oracle | asserts-oracle 0.29 | diverges 0.16 | 🔴 violation | escalate (violation) |
| BR-8 | Handed-over secrets are hidden in every text that reaches the developer. | No developer-visible text contains the handed-over secret | conversation_link.go:119; result_payload.go:63-70; claude_code_proxy.go:222-227 | result_payload_test.go:137; credential_test.go:10 | shallow | produces-oracle | asserts-oracle 0.75 | produces-oracle 0.44 | ✅ conforms | escalate (disagree) |
| BR-9 | Edge: Claude Code stops on its own → "the conversation ended" from then on; an answer that cannot be understood is ignored unless it is a final answer. | Unreadable non-final lines ignored; an unreadable final answer surfaces as a failure for that message (not silently dropped) | conversation_link.go:103-112; result_payload.go:28-43 | — (result_payload_test.go:117 covers decoding only) | no-test | diverges | asserts-oracle 0.25 | produces-oracle 0.47 | ✅ conforms | escalate (disagree) |
| NFR-1 | Performance: a message costs no process start; giving up returns within one second. | One process per conversation; give-up ≤1s | claude_code_proxy.go:132-149; conversation_link.go:131-156 | test/conversation_test.go:122 | shallow | produces-oracle | asserts-oracle 0.48 | produces-oracle 0.21 | ✅ conforms | escalate (disagree) |
| NFR-2 | Security: secrets never appear in failures; nothing is logged. | No failure carries the secret; no logging | conversation_link.go:119-126; claude_code_proxy.go:222-227 | credential_test.go:10; result_payload_test.go:137 | shallow | produces-oracle | asserts-oracle 0.32 | produces-oracle 0.71 | ✅ conforms | escalate (disagree) |
| NFR-3 | Compatibility: very long messages work on every platform. | Long messages travel on input, never as an argument | conversation_link.go:133-137; query_command.go:44-47 | query_command_test.go:55-77 | shallow | produces-oracle | no-test 0.41 | produces-oracle 0.55 | 🟡 partial | escalate (disagree) |
| NFR-4 | Concurrency: a conversation may be used from several goroutines; sends are serialised by refusal, interrupts may come from any goroutine. | Goroutine-safe; concurrent sends refused; interrupt from any goroutine | query/conversation.go:73-151; conversation_link.go:195-208 | query/conversation_test.go:115-163 | asserts-oracle | produces-oracle | asserts-oracle 0.73 | diverges 0.35 | 🔴 violation | escalate (violation) |
| NFR-5 | Dependencies: none beyond the standard library in production code. | Production imports stdlib only | `go list` imports of ./query, ./internal/infrastructure/claudecode | — | no-test | produces-oracle | no-test 1.00 | produces-oracle 0.81 | 🟡 partial | recheck (mid confidence) |

### Notes (shadow evidence, rechecks, disagreements)

- **BR-5 / AC-6 — stale answer handed to the next message (probe P2, reproduced).** `conversation_link.go:152-153`: when the 750 ms settle expires, Send calls `l.stop()` and returns, but the link is not marked ended and the one-slot `answers` buffer is not drained. The reader does a non-blocking send into that buffer (`:106-109`), so a final answer that arrives late is kept. `Conversation` resets to idle (`conversation.go:110`). The next `Send` writes successfully while the process is still dying, and its select picks up the buffered stale answer. Probe output: `B: reply="ESSAY (stale)" cost=0.02 err=<nil>` for "Reply with exactly: AFTER". Only the message after that fails as ended. Jev's "conforms" stands as the status, but Claude's shadow verdict is **diverges**.
- **BR-3 — spend of a given-up message is dropped (probe P3, reproduced).** `conversation_link.go:149-150` reads the settling answer and discards it, so its cost never reaches `Conversation.total`. Probe: after giving up, `TotalCostUSD()` is `0` although Claude Code stated 0.02. The next reply reports `CostUSD=0.03` when its own increase was 0.01. If the program closes after giving up, that spend is never reported. Shadow verdict: **diverges**.
- **BR-9 — unreadable final answer hangs the message (probe P1, reproduced).** `conversation_link.go:103-112`: a `type:"result"` line that fails `newResultPayload` (e.g. no `total_cost_usd`, which `result_payload_test.go:125` proves is a decode error) is valid JSON, so it is dropped silently. Send then waits until ctx ends: forever with `context.Background()`, or until the deadline in the probe (`after 3.751s ... context deadline exceeded`). One-shot `Ask` turns the same answer into `*UnreadableReplyError`. Shadow verdict: **diverges**.
- **BR-4 — Interrupt check-then-act race (from reading the code; did not reproduce in 20k probe runs).** `conversation.go:131-137` reads `state == answering` under the mutex, releases it, then writes the interrupt. A interrupt aimed at message A can reach Claude Code after A's answer was emitted, while Claude Code is idle (its behaviour there is unverified), or after message B was written, which aborts B. The idle-delivery window, from A's result line to Send reacquiring the mutex, is wide. Shadow verdict: **diverges**.
- **BR-6 — Close can hang / children not signalled.** Close takes the `writing` mutex (`conversation_link.go:182`). `Send` holds that mutex during a blocking `input.Write` (`:198-207`) that ignores ctx. If Claude Code stops reading input, both Send and Close hang (probe P5: still blocked 5 s after a 500 ms deadline). When Claude Code exits within 2 s of input closing, the library never signals the process group, and tool commands run in their own session, so ending "commands Claude started" relies entirely on Claude Code. The live test only pgreps the Claude Code process. Shadow: **unclear** code, **shallow** test.
- **BR-2** — `--max-turns` is passed once to the long-lived process. Whether Claude Code applies it per message in stream-json mode, and whether the process survives `error_max_turns`, cannot be seen from code, and no test asserts per-message behaviour.
- **BR-7** — Jev's violation pick is low confidence (0.16). Claude finds the six listed kinds mapped correctly (`domain.go:40-66`). The test gap is that no conversation test covers the `problem` kind. Recheck: agrees with Jev on shallow, disagrees on diverges.
- **NFR-4** — Jev's violation pick (0.35) is consistent with the BR-4 race and the P5 Close/Send mutex hang.
- **AC-1 / AC-12 / BR-1 (unclear by Jev)** — recheck: the paths are short and serialised by the conversation mutex, and the CLI-dependent parts (memory, `--resume`) are covered by live tests. Claude keeps produces-oracle.
- **AC-5** — the test asserts "no failure" and "link not called" but never sends the next message.
- **Rechecks AC-4, AC-7, AC-9, NFR-5** — re-read. The mapped unit tests pass under `-race`. Production imports: stdlib plus the module's own `auth`, `installation` and `query`.
- **P6 (minor)** — a Send whose ctx was already cancelled still writes the message (`conversation_link.go:137`) and then interrupts it, so Claude starts work (and usage) on a request the caller has already abandoned.

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| query/conversation.go:155-159 | `Conversation.SessionID()` exposes the conversation reference | undocumented (implied by US-05, no clause) |
| query/conversation.go:103-105 | A message in flight when Close is called reports "the conversation is closed" even if a reply arrived | undocumented (state diagram only) |
| query/conversation.go:88-90 | The empty check comes before the state check, so an empty message on a closed, ended or busy conversation reports "empty" | undocumented ordering |
| conversation_link.go:161-163 | `Interrupt(ctx)` with a finished ctx fails as given up | undocumented |
| conversation_link.go:27, 185-190 | Close waits up to 2 s for Claude Code to exit on its own before stopping it | undocumented timing |
| claude_code_proxy.go:131, 137 | Open detaches from ctx (`WithoutCancel`), so ctx does not bound starting despite the comment | undocumented / comment inaccurate |
| conversation_link.go:79-80 | stderr is gathered into an unbounded buffer for the whole conversation (unlike stdout diagnostics, which are capped at 4 KiB) | undocumented (memory growth in long conversations) |

No orphan implements an Out-of-Scope item: there is no partial text, no permission prompting and no parallel answering, and setting sources are not chosen.

## Summary

- Conforms: 22/30 clauses ✅ (73%), per Jev statuses. Claude's shadow verdict would count only 14 as conforming, and it finds BR-3, BR-4, BR-5 and BR-9 divergent.
- Violations: BR-7, NFR-4 (Jev picks; see Notes)
- Mis-asserted: AC-5
- Partial: NFR-3, NFR-5
- Gaps: none
- Unclear: AC-1, AC-12, BR-1
- Orphans: 7
- Jev: ran on 30 clauses
- Needs James: AC-1 (disagree), AC-3 (low confidence), AC-6 (low confidence), AC-8 (low confidence), AC-12 (disagree), BR-1 (disagree), BR-2 (disagree), BR-3 (disagree; probe P3 shows divergence), BR-4 (disagree; race), BR-5 (disagree; probe P2 shows divergence), BR-6 (disagree), BR-7 (violation), BR-8 (disagree), BR-9 (disagree; probe P1 shows divergence), NFR-1 (disagree), NFR-2 (disagree), NFR-3 (disagree), NFR-4 (violation)

Ceiling: static conformance audit against the PRD acceptance criteria. Throwaway probes (scratchpad only, `sh` stand-ins, never Claude Code) were used to confirm suspected defects. Integration tests were not run.

---

## Resolution (after the audit)

Each finding was judged against the rules; the matrix above is left as audited.

| Finding | Decision | Change |
| :--- | :--- | :--- |
| Defect 1 / BR-5 — a late answer reached the next message (probe P2) | Real defect — fixed | A link that cannot settle a given-up message within 0.6 s is marked broken and stopped; every later message fails as "the conversation ended". Unit test with a real, unread OS pipe |
| Defect 2 / BR-3 — the cost of a given-up message was lost (probe P3) | Real defect — fixed | The settling answer is returned with the give-up error and counted in the total; the next reply costs only its own increase |
| Defect 3 / BR-9 — an unreadable final answer waited forever (probe P1) | Real defect — fixed | Lines are classified by `type` first; a final answer that cannot be decoded is delivered as `*UnreadableReplyError` |
| Defect 4 — `Send`/`Close` hung on a write Claude Code stopped reading (probe P5) | Real defect — fixed | Writes run under the message's context; giving up breaks the link; `Close` no longer shares a lock with writes and closing the input releases a blocked write |
| Defect 5 / BR-4 — an interrupt could abandon a later message | Real race — fixed | Probed: Claude Code handles lines in order and ignores an interrupt while idle. `Interrupt` now writes while holding the conversation lock, so it always precedes the next message |
| P6 — an already-ended context still sent the message | Fixed | Checked before writing; unit test reads the pipe and finds nothing |
| Unbounded diagnostics buffer | Fixed | Kept to the first 4 KB |
| `Open` comment / ignored context | Fixed | An already-ended context refuses to start; the conversation lives until `Close` by design |
| AC-5 — idle interrupt test did not send the next message | Fixed | It now does |
| BR-2 — round cap per message unverified | Fixed | Live test: cap 1 stops the first message; the next is answered |
| BR-7 / NFR-3 — no "problem" or long-message test in conversations | Fixed | Free integration tests (unavailable model; 440 KB message) |
| Orphan 3 — empty check before state check | Fixed | State refusals come first |
| NFR-5 — no test pins standard-library-only | Already covered | `boundary_test.go` guards every production file, including these |
| Close and commands Claude started after a normal exit | Accepted | Closing the input lets Claude Code end its own commands; the group stop applies when it does not finish within 2 s |
| Orphans 1, 2, 4, 5 | Benign / documented | `SessionID()`, closing during a reply, `Interrupt` with an ended context, the 2 s close grace |
