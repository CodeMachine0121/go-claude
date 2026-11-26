# Contract Traceability Matrix — 2026-11-26-cli-auth-profiles

Contract: PRD.md (v1.0, Finalized)
Design map: ARCH.md
Implementation: root package `claude` (client.go, option.go), installation/, auth/, internal/infrastructure/claudecode/; integration tests test/auth_test.go (`//go:build integration`)
Oracle: Acceptance Criteria — 46 clauses (28 AC from Section 3, 9 Core Business Rules + 1 Edge Case from Section 4, 8 NFR from Section 6)
Audit date: 2026-11-26

Ceiling: static conformance audit. Test assertions and code paths were judged against oracles derived from the PRD before any code was opened. No new probes were written and no invented scenarios were executed. Corroboration was limited to four mapped unit tests (TestUseCase_Verify, TestVersionVo_AtLeast, TestSettings_Environment, TestNewStatusPayload_ToVo), all green. No integration test was run.

Out of Scope (negative checklist): sending requests / replies; multi-turn conversations; tool permissions / extensions; installing or updating Claude Code; any sign-in that bypasses Claude Code. No code implements any of these.

## Clauses

The `Spec-expected` column holds the business-observable oracle from Phase 2. The concrete artifact it bridges to (via UL-MAP/ARCH) is what the audit columns check.

`Test audit` / `Code audit` are Claude's shadow verdict. `Jev test` / `Jev code` are Jev's pick with its confidence. `Status` and `Route` are copied from `jev_audit.py` as returned.

Path prefix `cc/` = `internal/infrastructure/claudecode/`.

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev test | Jev code | Status | Route |
|----|--------|------------------------|------|------|------------|------------|----------|----------|--------|-------|
| AC-1 | US-01 Claude Code is found automatically — 2.1.332 set up, no chosen copy; prepare → succeeds, reports 2.1.332 | Preparation succeeds; library reports release 2.1.332 | client.go:22-38; cc/settings.go:35-40; cc/claude_code_proxy.go:42-55 | installation/usecase_test.go:21-25; test/auth_test.go:85 | asserts-oracle | produces-oracle | asserts-oracle 0.77 | unclear 0.32 | ❔ unclear | escalate (disagree) |
| AC-2 | US-01 The developer's chosen copy is used | The chosen copy is the one run | option.go:25; cc/settings.go:42-45 | test/auth_test.go:99; cc/settings_test.go:73 | asserts-oracle | produces-oracle | asserts-oracle 0.81 | produces-oracle 0.95 | ✅ conforms | recheck |
| AC-3 | US-01 Exactly the oldest supported release is accepted (2.1.0) | No refusal; reports 2.1.0 | installation/domain.go:23-28; installation/vo.go:37-45 | installation/usecase_test.go:26-30 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.80 | ✅ conforms | recheck |
| AC-4 | US-01 A release older than supported is refused (2.0.9) | Refused "too old"; names 2.0.9 found, 2.1.0 required | installation/domain.go:24-25; installation/errors.go:28-30 | installation/usecase_test.go:31-36, :74 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.97 | ✅ conforms | accept |
| AC-5 | US-01 Claude Code is missing | Refused "not installed"; tells how to get Claude Code | cc/settings.go:36-38; installation/errors.go:10 | installation/usecase_test.go:37-42; test/auth_test.go:112 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.95 | ✅ conforms | accept |
| AC-6 | US-01 The chosen copy does not exist (another is set up) | Refused "not installed"; other copy not used | cc/settings.go:42-44 | test/auth_test.go:122; cc/settings_test.go:80 | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.91 | ✅ conforms | accept |
| AC-7 | US-02 Signed in with a Claude subscription (team, expires 24 Dec 2026) | Signed in, kind "Claude subscription", plan "team", expiry 24 Dec 2026 | cc/status_payload.go:35-41, :50-51; auth/domain.go:16-23 | cc/status_payload_test.go:62-70; auth/usecase_test.go:25-35 | asserts-oracle | produces-oracle | asserts-oracle 0.67 | unclear 0.28 | ❔ unclear | escalate (disagree) |
| AC-8 | US-02 Signed in with a developer account | Signed in, kind "developer account" | cc/status_payload.go:52-53 | cc/status_payload_test.go:86-90; auth/usecase_test.go:36-40 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.89 | ✅ conforms | accept |
| AC-9 | US-02 Not signed in is an ordinary answer | Not signed in; no failure | cc/claude_code_proxy.go:60-69; auth/domain.go:13-15 | cc/status_payload_test.go:71-75; test/auth_test.go:150 | asserts-oracle | produces-oracle | asserts-oracle 0.81 | produces-oracle 0.76 | ✅ conforms | recheck |
| AC-10 | US-02 An unfamiliar account kind is kept | Signed in; kind = the name Claude Code gave | cc/status_payload.go:48 | cc/status_payload_test.go:91-95; auth/usecase_test.go:46-50 | asserts-oracle | produces-oracle | asserts-oracle 0.39 | produces-oracle 0.83 | ✅ conforms | escalate (low confidence) |
| AC-11 | US-02 Claude Code's answer cannot be understood | Fails "could not read sign-in state", including what Claude Code said | cc/claude_code_proxy.go:65-67; auth/errors.go:29-31 | auth/usecase_test.go:61-66; cc/status_payload_test.go:108 | shallow | produces-oracle | asserts-oracle 0.77 | produces-oracle 0.93 | ✅ conforms | escalate (disagree) |
| AC-12 | US-02 The developer gives up waiting (Claude Code not yet answered) | Check stops at once; reported as given up | cc/claude_code_proxy.go:107-121 | test/auth_test.go:275; auth/usecase_test.go:67-71 | shallow | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.65 | ✅ conforms | escalate (disagree) |
| AC-13 | US-03 Sign in with a Claude subscription by default | Subscription sign-in in the person's own terminal; afterwards signed in | auth/domain.go:37-40; cc/sign_in_command.go:13-14; option.go:71; auth/usecase.go:28-44 | auth/usecase_test.go:109-116; cc/sign_in_command_test.go:17; option_test.go:25 | asserts-oracle | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.85 | ✅ conforms | accept |
| AC-14 | US-03 Sign in with a developer account and a pre-filled email | Developer-account sign-in with dev@example.com pre-filled | cc/sign_in_command.go:15-22 | auth/usecase_test.go:117-124; cc/sign_in_command_test.go:22 | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.97 | ✅ conforms | accept |
| AC-15 | US-03 Sign in with company single sign-on | Single sign-on sign-in | cc/sign_in_command.go:17-18 | auth/usecase_test.go:125-132; cc/sign_in_command_test.go:27 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.99 | ✅ conforms | accept |
| AC-16 | US-03 The person does not finish signing in | Fails "sign-in was not completed" | cc/claude_code_proxy.go:78-80; auth/usecase.go:40-42 | auth/usecase_test.go:133-149 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.89 | ✅ conforms | accept |
| AC-17 | US-03 More than one sign-in method is chosen | Refused "choose one sign-in method"; Claude Code not asked | auth/domain.go:34-36; auth/usecase.go:29-32 | auth/usecase_test.go:150-155, :186-188 | asserts-oracle | produces-oracle | asserts-oracle 0.91 | produces-oracle 0.92 | ✅ conforms | accept |
| AC-18 | US-03 The developer gives up on the sign-in (sign-in has started) | Sign-in stopped; reported as given up | cc/claude_code_proxy.go:107-121 | test/auth_test.go:289; auth/usecase_test.go:156-162 | shallow | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.93 | ✅ conforms | escalate (disagree) |
| AC-19 | US-04 Sign out when signed in | Afterwards state reads not signed in | auth/usecase.go:47-49; cc/claude_code_proxy.go:85-94 | auth/usecase_test.go:212 | shallow | produces-oracle | shallow 0.64 | unclear 0.49 | ❔ unclear | escalate (disagree) |
| AC-20 | US-04 Sign out when already signed out | Succeeds; no failure | cc/claude_code_proxy.go:85-94 | test/auth_test.go:206 | asserts-oracle | produces-oracle | asserts-oracle 0.53 | produces-oracle 0.84 | ✅ conforms | escalate (low confidence) |
| AC-21 | US-05 No profile means the everyday sign-in | Reports everyday account A | cc/settings.go:51-73 | cc/settings_test.go:26-30; test/auth_test.go:164 | asserts-oracle | produces-oracle | asserts-oracle 0.88 | produces-oracle 0.61 | ✅ conforms | recheck |
| AC-22 | US-05 A profile has its own sign-in (work = B, everyday = A) | Profile reports B; everyday still A | option.go:90-99; cc/settings.go:53-55, :67-69 | test/auth_test.go:175; cc/settings_test.go:31-40 | shallow | produces-oracle | asserts-oracle 0.77 | produces-oracle 0.41 | ✅ conforms | escalate (disagree) |
| AC-23 | US-05 Signing out of a profile leaves the everyday sign-in alone | work not signed in; everyday still signed in | cc/settings.go:51-73; cc/claude_code_proxy.go:85-94 | test/auth_test.go:190 | shallow | produces-oracle | asserts-oracle 0.93 | unclear 0.21 | ❔ unclear | escalate (disagree) |
| AC-24 | US-05 A new profile starts signed out | Not signed in; no failure | cc/claude_code_proxy.go:60-69 | test/auth_test.go:150 | asserts-oracle | produces-oracle | asserts-oracle 0.92 | produces-oracle 0.53 | ✅ conforms | escalate (low confidence) |
| AC-25 | US-06 A subscription pass is used without an interactive sign-in | Signed in, kind "subscription pass"; no sign-in started | option.go:100-105; cc/settings.go:56-71; cc/status_payload.go:54-55 | test/auth_test.go:235-260 | asserts-oracle | produces-oracle | asserts-oracle 0.45 | produces-oracle 0.41 | ✅ conforms | escalate (low confidence) |
| AC-26 | US-06 A developer account key is used | Signed in, kind "developer account key" | option.go:106-111; cc/status_payload.go:56-57 | test/auth_test.go:235-260 | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.77 | ✅ conforms | recheck |
| AC-27 | US-06 Both kinds of credential are handed over | Refused "supply only one kind of credential" | option.go:80-82; client.go:23-26 | client_test.go:19-23; example_test.go:52 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.85 | ✅ conforms | accept |
| AC-28 | US-06 A failure never repeats the handed-over secret | Failure lacks the secret; the spot is marked hidden | cc/claude_code_proxy.go:124, :134; cc/credential.go:34-39 | cc/credential_test.go:10-42 | shallow | produces-oracle | asserts-oracle 0.40 | produces-oracle 0.48 | ✅ conforms | escalate (disagree) |
| BR-1 | Oldest supported release is 2.1.0; releases compared part by part | Major, then minor, then patch, compared numerically | installation/vo.go:37-45 | installation/vo_test.go:39-59 | asserts-oracle | produces-oracle | asserts-oracle 0.84 | produces-oracle 0.77 | ✅ conforms | recheck |
| BR-2 | A chosen copy is used as-is; when it does not exist, no other copy is tried | Chosen used verbatim; missing → not installed, no fallback | cc/settings.go:42-45 | cc/settings_test.go:73-86 | asserts-oracle | produces-oracle | asserts-oracle 0.92 | produces-oracle 0.85 | ✅ conforms | accept |
| BR-3 | "Not signed in" is a state, never a failure — for checking and signing out | Check and sign-out return no failure when signed out | cc/claude_code_proxy.go:57-94 | cc/status_payload_test.go:71; test/auth_test.go:150, :206 | asserts-oracle | produces-oracle | asserts-oracle 0.79 | produces-oracle 0.57 | ✅ conforms | escalate (low confidence) |
| BR-4 | Exactly one sign-in method per sign-in; Claude subscription is the default | Default subscription; combinations refused | auth/domain.go:33-42 | auth/usecase_test.go:109-155 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.87 | ✅ conforms | accept |
| BR-5 | At most one kind of credential may be handed over | Both → refused | option.go:80-82 | client_test.go:19-23 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.91 | ✅ conforms | accept |
| BR-6 | A profile isolates sign-in and conversation history; no profile means the everyday sign-in | Profile has separate sign-in AND history; none = everyday | cc/settings.go:51-73; option.go:90-99 | cc/settings_test.go:14-66; test/auth_test.go:175, :217 | shallow | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.81 | ✅ conforms | escalate (disagree) |
| BR-7 | A handed-over secret is hidden wherever it would otherwise appear in what the library reports | Nothing the library reports contains the secret | cc/claude_code_proxy.go:50 → installation/vo.go:23 (unredacted) | cc/credential_test.go:10-42 | shallow | diverges | asserts-oracle 0.53 | produces-oracle 0.33 | ✅ conforms | escalate (disagree) |
| BR-8 | Unrecognised account kinds are reported by the name Claude Code gave, not rejected | Raw name kept | cc/status_payload.go:48 | cc/status_payload_test.go:91-95 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.97 | ✅ conforms | accept |
| BR-9 | The library never reads saved sign-in secrets and never signs in on its own | No saved-secret read; sign-in only on request | cc/claude_code_proxy.go:41-94; auth/usecase.go:28-44 | — | no-test | produces-oracle | no-test 1.00 | produces-oracle 0.96 | 🟡 partial | accept |
| BR-10 | Edge case: Claude Code disappears between preparation and use → "Claude Code is not installed" | Later operations refused "not installed" | cc/settings.go:35-45; cc/claude_code_proxy.go:103-106 | test/auth_test.go:131 | asserts-oracle | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.91 | ✅ conforms | accept |
| NFR-1 | Checking sign-in adds no noticeable delay beyond Claude Code's own answer time | One Claude Code run + decode | cc/claude_code_proxy.go:60-69 | — | no-test | produces-oracle | no-test 1.00 | produces-oracle 0.52 | 🟡 partial | escalate (low confidence) |
| NFR-2 | Giving up stops an operation within one second | Stop ≤ 1 s after giving up mid-run | cc/claude_code_proxy.go:20, :107-109 | test/auth_test.go:275-321 | shallow | unclear | asserts-oracle 0.45 | produces-oracle 0.84 | ✅ conforms | escalate (disagree) |
| NFR-3 | Handed-over secrets never appear in failures, refusals or descriptions of the library's settings | No secret in any failure/refusal/settings print | cc/claude_code_proxy.go:50 → installation/vo.go:23 (unredacted); cc/credential.go:42-48 | cc/credential_test.go:10-55; client_test.go:36 | shallow | diverges | asserts-oracle 0.88 | produces-oracle 0.41 | ✅ conforms | escalate (disagree) |
| NFR-4 | The library never reads saved sign-in secrets and never performs its own sign-in | As BR-9 | cc/claude_code_proxy.go:41-94 | — | no-test | produces-oracle | no-test 1.00 | produces-oracle 0.73 | 🟡 partial | recheck |
| NFR-5 | Account identity details (email, organisation) available to the developer but never included in failures | StatusDto carries email/org; no failure text does | cc/claude_code_proxy.go:67; auth/errors.go:29-31; cc/status_payload.go:35-40 | cc/status_payload_test.go:115 (decode only) | no-test | diverges | shallow 0.56 | produces-oracle 0.25 | 🟠 mis-asserted | escalate (disagree) |
| NFR-6 | macOS, Linux, Windows wherever Claude Code 2.1.0+ runs; two most recent Go releases | Works on all three OSes and the two latest Go releases | go.mod:3; cc/settings.go:35-45 | — (no CI matrix) | no-test | produces-oracle | no-test 1.00 | unclear 0.20 | ❔ unclear | escalate (disagree) |
| NFR-7 | The library adds no third-party dependencies to the developer's program | No third-party package reaches the developer's program | production imports stdlib only; go.mod:5 requires testify | — | no-test | produces-oracle | no-test 1.00 | diverges 0.61 | 🔴 violation | escalate (violation) |
| NFR-8 | Analytics: none; the library collects and sends nothing | No collection / sending | no net/* import anywhere in production code | — | no-test | produces-oracle | no-test 1.00 | unclear 0.45 | ❔ unclear | escalate (disagree) |

### Notes (shadow reasoning, recheck findings)

- **AC-1:** I judged produces-oracle. `Installation().Version` comes from parsing `claude --version`, and the integration test compares it to the CLI's own first field. Jev's `unclear` probably reflects that 2.1.332 itself is only exercised through a mock.
- **AC-7:** `loginExpiresAt` is parsed as RFC3339 from a recorded 2.1.332 payload, and the unit test pins kind, plan "team" and expiry 2026-12-24. I see no divergence. Jev's `unclear` is likely about whether a real CLI always emits that field.
- **AC-11 (shadow: shallow):** the use-case test builds the `UnreadableStatusError{Output:"garbage"}` inside the mock, so the "includes what Claude Code said" assertion is tautological. The real wrapping at cc/claude_code_proxy.go:65-67 is untested. `TestNewStatusPayload_Unreadable` only asserts `err != nil`. Under `.claude/rules/testing.md` (no fake executables, no unit tests of proxy process I/O), testing that wrapping would mean moving it into a pure function.
- **AC-12 / AC-18 / NFR-2 (shadow: shallow):** `TestGivingUpStopsTheCheck`, `TestGivingUpStopsTheSignIn` and `TestGivingUpStopsTheSignOut` cancel the context **before** calling. `exec.Cmd.Start` then refuses at once and no Claude Code process ever starts. The "has not yet answered" / "a sign-in has started" precondition is never set up, so mid-run kill and `WaitDelay` behaviour (and the < 1 s bound) are not exercised. The use-case tests only pass a mock-built `context.Canceled` through. For NFR-2 the code sets `WaitDelay = 1s` (cc/claude_code_proxy.go:20,109). If a grandchild process keeps the output pipes open, the worst case is about 1 s plus kill latency, which leaves no margin against "within one second". Hence my shadow verdict of unclear.
- **AC-19 (shallow):** the only test is a mock Logout returning nil. Nothing checks "afterwards not signed in" from a signed-in state. That is not achievable unattended without a real account, so it would need a manual or opt-in test.
- **AC-22 / AC-23 (shallow):** the Given ("profile work signed in as B") is never set up. Both tests use a fresh, signed-out profile. AC-23's sign-out is therefore a no-op, and no test proves that a real profile sign-out leaves the everyday sign-in alone. The tests check only the signed-in booleans, not account identity A vs B.
- **AC-28 (shallow):** `Credential.Redact` is unit-tested, but no test shows that a failure reaching the developer is redacted. The wiring at cc/claude_code_proxy.go:124,134 is untested.
- **BR-6 (shallow):** the conversation-history half of the rule is untested. The recorded payload shows `projectsDirectory` living under `configDirectory`, which is indirect evidence only.
- **BR-7 / NFR-3 (shadow: diverges):** `Inspect` passes **raw, unredacted** stdout (`string(run.output)`, cc/claude_code_proxy.go:50) to `installation.NewVersionVo`. On a non-version answer, that function embeds the text verbatim in the refusal (`fmt.Errorf("%w: %q", ErrUnrecognizedVersion, text)`, installation/vo.go:23). If `claude --version` exits 0 but prints text containing the supplied secret, `New` returns a refusal containing the secret. Every other error path uses the redacted `said`. The path is narrow but real. Jev disagreed (produces-oracle 0.33 / 0.41), so this is escalated.
- **NFR-5 (shadow: diverges):** `UnreadableStatusError.Output` is the whole Claude Code output (cc/claude_code_proxy.go:67) and is printed by `Error()` (auth/errors.go:30). The status JSON carries `email` and `orgName`. Any otherwise-valid answer that fails decoding, for example an expiry that is not RFC3339 (the case exercised at cc/status_payload_test.go:115), therefore puts the person's email and organisation into a failure. This conflicts with AC-11 ("include what Claude Code said"), so the spec needs a decision: scrub identity fields from the echoed output, or relax one clause. The `ErrSignOutFailed` text (cc/claude_code_proxy.go:91) also echoes raw output.
- **NFR-7:** production code imports only the standard library. Jev's `diverges` most likely reflects `require github.com/stretchr/testify v1.10.0` in go.mod:5. Module-graph pruning keeps it out of consumers' builds, but it does appear in their module graph and go.sum. Escalated as returned.
- **Recheck routes (AC-2, AC-3, AC-9, AC-21, AC-26, BR-1, NFR-4):** I re-read the paths. Mapped unit tests TestUseCase_Verify, TestVersionVo_AtLeast, TestSettings_Environment and TestNewStatusPayload_ToVo pass. Nothing changes.

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| installation/errors.go:18; installation/vo.go:20-34 | `ErrUnrecognizedVersion`: refusal when `claude --version` output has no major.minor.patch | undocumented (and the leak site in BR-7/NFR-3) |
| option.go:16-17, :85-110 | `ErrInvalidOption`: empty CLI path / profile / token / key, and an unresolvable relative profile dir, are refused | undocumented |
| auth/errors.go:19-20; cc/claude_code_proxy.go:90-92 | `ErrSignOutFailed`: sign-out failure when Claude Code exits non-zero; echoes Claude Code's raw text | undocumented |
| cc/claude_code_proxy.go:47-49 | `claude --version` exiting non-zero → plain `errors.New`-style "claude code failed: …", not a recognisable domain error (also against .claude/rules/code-style.md "wrap into errors the domain recognizes") | undocumented |
| option.go:49-57 | `WithTerminal` option redirects the interactive sign-in streams | undocumented (supports AC-13 but not in contract) |
| auth/domain.go:33-41; cc/sign_in_command.go:12-19 | A single undefined `SignInMethod` bit (e.g. `SignInMethod(8)`) passes validation and runs `auth login` with no method flag | undocumented |

No orphan matches an Out-of-Scope item.

## Summary

- Conforms: 35/46 clauses ✅ (76%)
- Violations: NFR-7 (Jev: testify in go.mod)
- Mis-asserted: NFR-5
- Partial: BR-9, NFR-1, NFR-4
- Gaps: none
- Unclear: AC-1, AC-7, AC-19, AC-23, NFR-6, NFR-8
- Orphans: 6
- Jev: ran on 46 clauses
- Needs James (23):
  - disagree: AC-1, AC-7, AC-11, AC-12, AC-18, AC-19, AC-22, AC-23, AC-28, BR-6, BR-7, NFR-2, NFR-3, NFR-5, NFR-6, NFR-8
  - low confidence: AC-10, AC-20, AC-24, AC-25, BR-3, NFR-1
  - violation: NFR-7
- Highest priority (shadow audit found code divergence that Jev's `conforms` / `mis-asserted` masks):
  - BR-7 / NFR-3: secret leak through the unredacted version refusal
  - NFR-5: email/org leak through `UnreadableStatusError.Output`, in tension with AC-11

---

## Resolution (after the audit)

Each finding was judged against the rules; the matrix above is left as audited.

| Finding | Decision | Change |
| :--- | :--- | :--- |
| BR-7 / NFR-3 — release text reached refusals unredacted | Real leak — fixed | Every text Claude Code prints is redacted inside the single execution path, so no caller can forget |
| NFR-5 — unreadable answer echoed email / organisation | Real leak; conflicted with AC-11 — fixed | Identity fields are removed before the answer is shown; AC-11 reworded to say so |
| Orphan 6 — undefined sign-in method passed validation | Bug — fixed | Refused as "unknown sign-in method"; added as a business rule |
| Orphan 4 — release failure returned an untyped error | Rule breach — fixed | Wrapped as "Claude Code release is not recognized"; added as an edge case |
| NFR-2 — stop delay exactly at the one-second limit | Fixed | Given-up runs are stopped within half a second |
| NFR-7 — test tool listed in the module file | Not a violation: test-only tools are never compiled into the developer's program | NFR reworded; a guard test now fails if production code imports third-party code |
| NFR-8 / BR-9 / NFR-4 — no test | Fixed | Guard tests: production code never imports networking, never refers to stored sign-in secrets |
| NFR-6 — no platform matrix | Fixed | CI on Linux, macOS and Windows with the oldest and newest Go |
| Orphans 2, 3, 5 — invalid options, sign-out failure, sign-in streams | Correct behavior missing from the spec | Added to the PRD business rules / edge cases |
| AC-11, AC-28 — wrapping and redaction inside the process path not unit-tested | Partly fixed | Identity removal is a pure, tested type; redaction now happens once in the execution path. Driving a failing Claude Code still needs a fake executable, which the rules forbid |
| AC-12 / AC-18 — give-up tests cancel before the process starts | Accepted limitation | A long-running real command that can be cancelled safely does not exist unattended (sign-in opens a browser) |
| AC-19 / AC-22 / AC-23 / BR-6 — need a profile signed in to a real second account | Accepted limitation | Cannot run unattended; verify manually with `examples/signin -profile` |
| NFR-1 — no timing test | Accepted | One Claude Code run per check; no library work worth measuring |
