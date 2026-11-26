# Conversation — Requirements Brief

## Goal
Let a developer keep a conversation with Claude open and exchange several messages in it, each answered with one complete final reply, without starting over every time. While Claude is answering, the developer can interrupt it and keep talking afterwards. Each reply says what it cost on its own, and the conversation keeps a running total. Closing the conversation ends Claude Code's work cleanly.

## Requirements
- Open a conversation, optionally continuing an earlier one, with the same choices as a one-shot request: model, project folder, actions allowed without asking, and a round cap that applies to each message.
- Send a message and receive Claude's final reply; Claude remembers the earlier messages of the conversation.
- Only one message is answered at a time; sending another while Claude is still answering is refused as "still answering".
- Interrupt Claude while it answers: the reply being written is abandoned and reported as "interrupted"; the conversation stays open and the next message is answered normally. Interrupting when Claude is not answering does nothing.
- Giving up on one message stops that reply only; the conversation stays usable.
- Each reply reports the rounds used and what that reply cost; the conversation reports the total cost so far.
- Closing the conversation ends Claude Code's work; closing twice is harmless; sending after closing is refused as "the conversation is closed".
- If Claude Code stops on its own, the waiting message fails as "the conversation ended" and so does every later one.
- The same failures as one-shot requests apply per message: empty message, round limit reached, not signed in, a problem Claude explains, an unknown earlier conversation, a work folder that does not exist.
- A handed-over pass or key never appears in any failure.

## Examples (Specification by Example)
Each example lists **only** the data that affects the behavior — nothing more.

### Rule: Talking in one conversation
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | An open conversation | Developer sends "Reply with exactly: ONE", then "What did you just reply? One word." | First reply "ONE"; second reply mentions "ONE" |
| 2 (exception) | An open conversation | Developer sends a message that is only spaces | Refused: "the request is empty"; the conversation stays open |
| 3 (exception) | Claude is still answering a message | Developer sends another message | Refused: "still answering"; the first reply still arrives |

### Rule: Interrupting
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Claude is writing a long essay | Developer interrupts | The essay message fails as "interrupted"; the next message "Reply with exactly: AFTER" gets "AFTER" |
| 2 (boundary)  | Claude is not answering anything | Developer interrupts | Nothing happens; no failure |
| 3 (exception) | Claude is writing a long essay | Developer gives up on that message | It is reported as given up within one second; the next message is answered normally |

### Rule: Costs
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | The conversation had cost 0.0038563 dollars; Claude Code now states a total of 0.00612419 | The second reply arrives | That reply cost 0.00226789 dollars; the total so far is 0.00612419 dollars |
| 2 (boundary)  | First reply of a conversation; Claude Code states a total of 0.0038563 | The reply arrives | That reply cost 0.0038563 dollars; the total so far is 0.0038563 dollars |

### Rule: Ending
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | An open conversation | Developer closes it | Claude Code's work ends; closing again does nothing |
| 2 (exception) | A closed conversation | Developer sends a message | Refused: "the conversation is closed" |
| 3 (exception) | Claude Code stopped on its own | Developer sends a message | Fails as "the conversation ended"; every later message fails the same way |

### Rule: Continuing an earlier conversation
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | An earlier conversation in which Claude replied "OK" | Developer opens it again and asks "What did you just reply?" | The reply mentions "OK" |
| 2 (exception) | A conversation that never took place | Developer opens it and sends a message | Fails as "conversation not found" |

## Out of Scope
- Seeing partial text while Claude writes (final replies only, as in one-shot requests).
- Asking the person for permission during a reply.
- Several messages answered at the same time in one conversation.
- Choosing which of the person's own Claude Code settings apply; the conversation uses them as Claude Code does.

## Open Decisions
- Whether the conversation belongs with one-shot requests or stands alone.

## Context / Background
- Third and last planned iteration. The developer asked to run the whole workflow at once.
- Observed with Claude Code 2.1.332: messages and replies flow over one open connection; every reply ends with a final answer whose cost is the running total of the conversation; an interrupt is acknowledged and the interrupted reply ends as aborted; the conversation then continues; the person's own Claude Code settings (including their prompt hooks) apply.
