# Agent-Team Request Options — Requirements Brief

## Goal
Let a developer run a team of Claude agents that share one Claude subscription, each agent with its own role, its own outside services and its own forbidden actions. When the shared subscription runs out of usage, every request and every conversation message must say so plainly — and say when the usage comes back whenever Claude Code states it — so a scheduler can pause until then instead of retrying in vain.

## Requirements
- A request refused because the subscription's usage limit is reached fails as "usage limit reached", distinct from every other problem.
- That failure carries the moment the usage comes back when Claude Code states one, and says nothing about it when Claude Code does not.
- The failure keeps the conversation reference, so the work can be continued after the limit resets.
- The same applies to every message of an open conversation; the conversation stays open and usable afterwards.
- A request or conversation can hand Claude written lists of the outside services it may connect to, and can ask Claude to use only those lists, ignoring the outside services the person set up for themselves.
- A list that cannot be found is refused before anything starts; a list that cannot be read fails as "invalid outside-service list".
- A request or conversation can list actions Claude must never take, even when they would otherwise be allowed.
- A request or conversation can add role instructions on top of Claude's usual instructions.
- All four choices apply equally to one-shot requests and to conversations, and leaving them out changes nothing.

## Examples (Specification by Example)
Each example lists **only** the data that affects the behavior — nothing more.

### Rule: Usage limit reached
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | The subscription's session usage is used up; Claude Code states it comes back at 15:00 UTC on 27 Nov 2026 | Developer sends a one-shot request | Fails as "usage limit reached", comes back at 15:00 UTC on 27 Nov 2026 |
| 2 (boundary)  | Usage is used up; Claude Code does not state when it comes back | Developer sends a one-shot request | Fails as "usage limit reached"; no moment is given |
| 3 (happy)     | An open conversation; usage is used up; Claude Code states 15:00 UTC | Developer sends a message | That message fails as "usage limit reached", 15:00 UTC; the conversation stays open |
| 4 (exception) | Claude reports some other problem, such as an unavailable model | Developer sends a request | Fails as "Claude reported a problem", never as "usage limit reached" |
| 5 (boundary)  | Usage is close to the limit but not used up | Developer sends a request | Claude answers normally |

### Rule: Outside services
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | An empty list of outside services; only the given lists are to be used | Developer sends a request | Claude connects to no outside service, not even the person's own |
| 2 (boundary)  | A list naming one service; the person's own services are not excluded | Developer sends a request | Claude can use the listed service and the person's own |
| 3 (exception) | A list that cannot be found | Developer sends a request or opens a conversation | Refused before starting: "invalid outside-service list" |
| 4 (exception) | A list whose content cannot be read | Developer sends a request | Fails as "invalid outside-service list" |
| 5 (boundary)  | A list named by its place inside the project folder | Developer sends a request | The list is found inside that project folder |

### Rule: Forbidden actions
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Editing and writing files are forbidden | Developer sends a request | Claude has no way to edit or write files |
| 2 (boundary)  | Nothing is forbidden | Developer sends a request | Claude has every action it would normally have |

### Rule: Role instructions
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Role instructions say "The codeword is PELICAN." | Developer asks "What is the codeword?" | Claude replies "PELICAN" |
| 2 (boundary)  | No role instructions | Developer sends a request | Claude follows its usual instructions only |
| 3 (boundary)  | An earlier conversation began with role instructions A; the developer continues it with role instructions B | Developer sends a message | Claude keeps following A, the instructions it began with |

## Out of Scope
- Waiting for the reset or retrying automatically — the scheduler decides.
- Handing over an outside-service list as words in the request instead of as a written list.
- Replacing Claude's usual instructions entirely.
- Reporting how much usage remains before the limit is reached.

## Open Decisions
None — decided with best practice (recorded under Context / Background).

## Context / Background
- Requested to drive agent teams on a new platform: one container per team, each agent in its own work folder with its own role and outside services, all sharing one subscription.
- Jev sized this as large (0.48); stages: clarify → prd → architecture → implement (code-first) → improve-codebase → contract.
- Observed with Claude Code 2.1.332: a refused request ends with a final answer marked as an error whose kind names the usage limit; the moment it comes back is stated only in the detailed answer stream, not in the short final answer; a missing or unreadable outside-service list stops Claude Code before any request is sent; forbidden actions disappear from Claude's tool list; role instructions are honoured.
- Decisions taken autonomously: one failure for lists that cannot be found or read; lists named by their place in the project are found from the project folder; the usage-limit failure keeps conversation, rounds and cost like the round-limit failure.
