# One-shot Request — Requirements Brief

## Goal
Let a developer hand Claude a single request through the person's Claude Code and get back one complete, final reply. Claude may work on the request over several rounds (thinking, using the tools the developer allows, looking at what they produce) before it answers; the developer can cap how many rounds it may take. Only the final reply is returned — no partial drafts. The reply comes with what the developer needs to account for it and to come back to it later: how many rounds were used, what it cost, and a conversation reference that lets a later one-shot request continue the same conversation.

## Requirements
- Send one request in plain words and receive Claude's final reply as text.
- Let the developer cap the number of rounds Claude may take; without a cap, Claude Code's own limit applies.
- When the cap is reached before Claude finishes, report it as "round limit reached", together with the rounds used, the cost so far and the conversation reference, so the developer can decide to continue.
- Let the developer choose which Claude model answers, which project folder Claude works in, and which kinds of action (such as running commands or editing files) Claude may take without asking; otherwise Claude Code's usual choices apply.
- Report the rounds used, the cost in US dollars (to the billionth of a dollar), and the conversation reference with every reply.
- Let a later one-shot request continue a previous conversation by its reference; an unknown reference is refused as "conversation not found".
- An empty request is refused before anything is sent.
- When the person is not signed in, the request fails as "not signed in".
- When Claude reports a problem (for example a model that is not available), the request fails with Claude's own explanation.
- Every request runs under the profile and the pass or key chosen when the library was prepared, and can be given up at any moment.
- A handed-over pass or key never appears in any failure.

## Examples (Specification by Example)
Each example lists **only** the data that affects the behavior — nothing more.

### Rule: Getting a final reply
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Request "Reply with exactly: OK" | Developer sends it | Reply "OK"; rounds used 1; a cost is reported; a conversation reference is reported |
| 2 (boundary)  | Request that is only spaces | Developer sends it | Refused before sending: "the request is empty" |

### Rule: Capping the rounds
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Cap 5; Claude finishes in 3 rounds | Developer sends the request | Final reply returned; rounds used 3 |
| 2 (boundary)  | Cap 1; Claude has to run a command first, so it needs a second round | Developer sends the request | Fails as "round limit reached"; rounds used, cost and conversation reference are reported |
| 3 (boundary)  | No cap given (a cap of zero counts as none) | Developer sends the request | Claude keeps working until it finishes, as Claude Code normally allows |
| 4 (exception) | A negative cap | Developer sends the request | Refused before sending: "the round limit cannot be negative" |

### Rule: Choosing how Claude works
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | The fastest Claude model chosen | Developer sends the request | That model answers |
| 2 (happy)     | The person's project folder chosen | Developer sends the request | Claude works inside that project folder |
| 3 (happy)     | Running commands allowed up front | Claude decides to run a command | It runs without asking the person |
| 4 (exception) | A project folder that does not exist | Developer sends the request | Refused before sending: "the work folder does not exist" |
| 5 (exception) | A model name Claude does not offer | Developer sends the request | Fails with Claude's explanation that the model is not available |

### Rule: Continuing a conversation
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | An earlier conversation in which Claude replied "OK" | Developer asks "What did you just reply?" in that same conversation | Claude's reply mentions "OK"; it is still the same conversation |
| 2 (exception) | A conversation that never took place | Developer tries to continue it | Fails as "conversation not found" |

### Rule: Accounting
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Claude Code says the request cost 0.00559738 dollars | Reply arrives | Cost reads 0.00559738 dollars |
| 2 (boundary)  | Claude Code says the request cost 0.0055135800000000006 dollars (digits beyond the billionth) | Reply arrives | Cost reads 0.00551358 dollars: anything beyond the billionth of a dollar is dropped |

### Rule: Sign-in and giving up
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (exception) | Person not signed in, no credential supplied | Developer sends a request | Fails as "not signed in" |
| 2 (exception) | Claude is still working | Developer gives up | Claude Code stops right away; reported as given up |
| 3 (exception) | A subscription pass was handed over; Claude's explanation of a problem repeats the pass | Failure reaches the developer | The pass does not appear in the failure; the spot is marked as hidden |

## Out of Scope
- Seeing intermediate drafts or streaming partial text (the developer chose final-only).
- A library-driven draft-review-improve loop.
- Multi-turn conversations kept open on one connection, and interrupting them (next iteration).
- Asking the person for permission to use a tool while Claude works; tools are either allowed up front or not used.
- Custom instructions for Claude, extensions and outside tool servers.

## Open Decisions
Items the PRD author should resolve:
- How the request is handed to Claude Code so that very long requests work on every platform.
- How "not signed in" is recognised.

## Context / Background
- This is the second of three iterations. Builds on the first (sign-in, profiles, credentials): every request runs through the same Claude Code runner.
- The developer chose: one-shot and multi-turn are separate; the round cap is Claude Code's own agent loop (not a library-driven refinement loop); only the final reply is returned.
- Observed with Claude Code 2.1.332: the final answer reports rounds used, cost, conversation reference and a stop reason; reaching the cap ends with "round limit reached"; an unknown model and "not signed in" come back as Claude-reported problems with an explanation; an unknown conversation reference is rejected with "No conversation found".
- Jev sizing skipped; stages as requested: clarify → prd → architecture → implement (code-first) → improve-codebase → contract.
