# Claude Code Sign-in and Profiles — Requirements Brief

## Goal
Give developers who build with this open-source library a dependable way to manage how their program is signed in to Claude, by relying on the Claude Code tool the person already has installed. A developer must be able to confirm Claude Code is available, see whether the person is signed in and with which kind of account, start a sign-in or sign-out on the person's behalf, keep several independent sign-in profiles side by side, and supply ready-made access credentials for unattended use. The library never handles the person's stored sign-in secrets itself; every sign-in goes through Claude Code, so whichever account Claude Code uses (including a Claude subscription) is the account the program uses.

## Requirements
- Find the person's Claude Code (or a particular copy the developer chooses) and refuse, with a plain explanation, when it is missing or too old.
- Tell the developer whether the person is signed in, with which kind of account (Claude subscription or pay-per-use developer account), on which plan, and until when.
- Treat "not signed in" as an ordinary answer, never as a failure.
- Let the person sign in by themselves, guided by Claude Code, choosing a Claude subscription (the default), a pay-per-use developer account, or their company's sign-on, with their email optionally filled in ahead of time.
- Let the person sign out; signing out when already signed out is quietly accepted.
- Let the developer keep several named profiles; each profile has its own sign-in and its own conversation history, separate from the person's everyday sign-in and from every other profile.
- Let the developer hand over a ready-made long-lived subscription pass or a developer account key so the program can run unattended; handing over both at once is refused as unclear.
- Never repeat a handed-over pass or key back in any message the library shows.
- Let the developer give up on any request at any moment; the library then stops right away and says the request was given up.
- Never look at the person's saved sign-in secrets and never sign in on its own; every sign-in goes through Claude Code.

## Examples (Specification by Example)
Each example lists **only** the data that affects the behavior — nothing more.

### Rule: Locating Claude Code
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Claude Code is set up on the person's computer, a recent release | Developer asks to use Claude Code | It is found and ready |
| 2 (happy)     | Developer points to one particular copy of Claude Code that exists | Developer asks to use Claude Code | That copy is used, even if the person has another copy set up |
| 3 (boundary)  | Claude Code is exactly the oldest supported release | Developer asks to use Claude Code | It is accepted |
| 4 (exception) | Claude Code is a release older than the oldest supported one | Developer asks to use Claude Code | Refused: "Claude Code is too old", naming the release found and the release required |
| 5 (exception) | Claude Code is not set up on the person's computer | Developer asks to use Claude Code | Refused: "Claude Code is not installed", with a pointer to how to get it |
| 6 (exception) | Developer points to a copy of Claude Code that does not exist | Developer asks to use Claude Code | Refused: "Claude Code is not installed" there; no other copy is tried instead |

### Rule: Reporting sign-in state
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Person signed in with a Claude subscription, team plan, expires 24 Dec | Developer checks sign-in state | Signed in; kind: Claude subscription; plan: team; expires 24 Dec |
| 2 (happy)     | Person signed in with a pay-as-you-go developer account | Developer checks sign-in state | Signed in; kind: developer account |
| 3 (boundary)  | Person never signed in, or signed out | Developer checks sign-in state | Not signed in — reported as a normal answer, not a failure |
| 4 (exception) | Claude Code answers with something the library cannot understand | Developer checks sign-in state | Failure: "could not read sign-in state", including what Claude Code said |
| 5 (exception) | Developer gives up waiting before Claude Code answers | Developer checks sign-in state | Stops promptly; reported as abandoned |

### Rule: Signing in
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | No method chosen | Developer starts sign-in | Interactive sign-in with a Claude subscription opens in the person's terminal and browser; afterwards the state reads signed in |
| 2 (happy)     | Method: developer account; email pre-filled | Developer starts sign-in | Developer-account sign-in opens with the email already filled in |
| 3 (happy)     | Method: company single sign-on | Developer starts sign-in | Single sign-on sign-in opens |
| 4 (exception) | Person closes the sign-in without finishing | Developer starts sign-in | Failure: "sign-in was not completed" |
| 5 (exception) | Two methods chosen at once | Developer starts sign-in | Refused before anything opens: "choose one sign-in method" |
| 6 (exception) | Developer gives up waiting | Developer starts sign-in | Sign-in is stopped; reported as abandoned |

### Rule: Signing out
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Person signed in | Developer signs out | Signed out; state now reads not signed in |
| 2 (boundary)  | Person already signed out | Developer signs out | Succeeds quietly, no failure |

### Rule: Separate profiles
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | No profile chosen; person signed in everyday as account A | Developer checks sign-in state | Account A's state is reported |
| 2 (happy)     | Profile "work" signed in as account B; everyday sign-in is account A | Developer checks state under profile "work" | Account B is reported; account A untouched |
| 3 (happy)     | Profile "work" signed in | Developer signs out under profile "work" | Only "work" is signed out; everyday sign-in stays signed in |
| 4 (boundary)  | Profile "fresh" has never been used | Developer checks state under profile "fresh" | Not signed in; the profile is ready to be signed in to |

### Rule: Ready-made access credentials
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy)     | Long-lived subscription access pass supplied; no interactive sign-in | Program uses Claude | Works under that subscription without opening any sign-in |
| 2 (happy)     | Developer account key supplied | Program uses Claude | Works under that developer account, billed per use |
| 3 (exception) | Both an access pass and a developer account key supplied | Developer sets up the program | Refused: "supply only one kind of credential" |
| 4 (exception) | Supplied access pass is rejected by Claude | Program uses Claude and the failure is shown | The failure message never contains the access pass itself |

## Out of Scope
- Sending requests to Claude and reading replies (next iteration).
- Multi-turn conversations, interrupting, and resuming conversations (later iteration).
- Tool permissions, extensions, and custom behaviors of Claude Code.
- Installing or updating Claude Code on the person's behalf.
- Any sign-in that bypasses Claude Code.

## Open Decisions
Items the PRD author should resolve:
- The exact oldest supported Claude Code release.
- Whether sign-in state also exposes account identity (email, organization) to the developer, and how to keep it out of diagnostic output.
- How "sign-in was not completed" is distinguished from other sign-in failures.

## Context / Background
- First of three planned iterations of an open-source Go library; developers add it to their own projects. The person running the program is assumed to have Claude Code installed.
- All clarification questions were resolved with best-practice defaults at the developer's request.
- Jev sized this as large (1.00); stages chosen: clarify → prd → architecture → implement (code-first) → improve-codebase → contract.
- Compliance note to carry into the public documentation: using a Claude subscription is fine for a person on their own machine; a product that offers Claude account sign-in to its own end users needs prior approval from Anthropic, otherwise it should use a developer account key.
