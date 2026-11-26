# 📔 Ubiquitous Language Map

**Project:** go-claude
**Bounded Context:** Using Claude from Go programs through the locally installed Claude Code
**Maintainer:** CodeMachine0121
**Last Updated:** 2026-11-26 (one-shot request added)

---

## 1. Nouns & Concepts

| Domain Term | Technical Name | User-Facing Label | Definition & Business Rules | Status |
| :--- | :--- | :--- | :--- | :--- |
| Developer | — | developer | The person writing a Go program that uses this library. | Confirmed |
| Person | — | user | The person running the developer's program; owns the Claude sign-in. | Confirmed |
| Claude Code | `claudecode.ClaudeCodeProxy` | Claude Code | The Claude Code tool installed on the person's computer. Every interaction with Claude goes through it. | Confirmed |
| Claude Code copy | `claude.WithCLIPath` | Claude Code path | A particular installation of Claude Code the developer points to, instead of the one found automatically. | Confirmed |
| Claude Code release | `installation.VersionVo` | version | The release number of Claude Code. Must be at least the oldest supported release. | Confirmed |
| Oldest supported release | `installation.MinimumSupportedVersion` | minimum version | The lowest Claude Code release the library accepts. | Confirmed |
| Sign-in state | `auth.StatusDto` / `auth.SignInStatusDomain` / `auth.StatusVo` | sign-in status | Whether the person is signed in, the account kind, the plan, and when the sign-in expires. "Not signed in" is an ordinary state. | Confirmed |
| Account kind | `auth.AccountKind` | account type | How the person is signed in: Claude subscription, developer account, ready-made subscription pass, developer account key, or not signed in. Unknown kinds are kept, not rejected. | Confirmed |
| Plan | `auth.StatusDto.Plan` | plan | The subscription tier (e.g. pro, max, team) when signed in with a Claude subscription. | Confirmed |
| Sign-in expiry | `auth.StatusDto.ExpiresAt` | expires at | When the current sign-in stops being valid; absent when it does not expire. | Confirmed |
| Sign-in method | `auth.SignInMethod` | sign-in method | How an interactive sign-in is done: Claude subscription (default), developer account, or company single sign-on. Exactly one per sign-in. | Confirmed |
| Profile | `claude.WithProfile` | profile | A named, isolated place holding its own sign-in and conversation history, separate from the everyday sign-in and every other profile. | Confirmed |
| Everyday sign-in | — | default profile | The sign-in Claude Code uses when no profile is chosen. | Confirmed |
| Subscription pass | `claude.WithOAuthToken` | OAuth token | A ready-made long-lived pass issued for a Claude subscription, for unattended use. Secret. | Confirmed |
| Developer account key | `claude.WithAPIKey` | API key | A key of a pay-per-use developer account, for unattended use. Secret. | Confirmed |
| One-shot request | `query.Request` | prompt | A single request handed to Claude that ends with one final reply. Must not be empty. | Confirmed |
| Final reply | `query.ResultDto.Text` | result | The complete answer Claude gives at the end; drafts in between are never returned. | Confirmed |
| Round | `query.ResultDto.Turns` | turn | One step of Claude's work on a request (think, act with a tool, look at the outcome). | Confirmed |
| Round limit | `query.Request.MaxTurns` | max turns | The most rounds Claude may take; 0 means no limit (Claude Code's usual behaviour); negative is refused. | Confirmed |
| Conversation reference | `query.ResultDto.SessionID` / `query.Request.SessionID` | session ID | Identifies a conversation so a later one-shot request can continue it. | Confirmed |
| Cost | `query.CostVo` | total cost (USD) | What Claude Code says the request cost, in US dollars, kept to the billionth of a dollar. Never a floating-point number. | Confirmed |
| Work folder | `query.Request.WorkingDirectory` | working directory | The project folder Claude works in; must exist. | Confirmed |
| Allowed action | `query.Request.AllowedTools` | allowed tools | A kind of action (running commands, editing files…) Claude may take without asking. | Confirmed |
| Model | `query.Request.Model` | model | Which Claude model answers; Claude Code's choice when absent. | Confirmed |
| Credential | `claudecode.Credential` | — | Either a subscription pass or a developer account key; at most one may be supplied. | Confirmed |

---

## 2. Actions & Processes

| Business Action | Technical Method | Trigger | Business Impact | Notes |
| :--- | :--- | :--- | :--- | :--- |
| Locate Claude Code | `installation.UseCase.Verify` (inside `claude.New`) | Developer creates the library entry point | Ready, or refused as not installed / too old | A chosen copy is never replaced by another |
| Check sign-in state | `auth.UseCase.Status` | Developer asks | Reports the sign-in state; no change | |
| Sign in | `auth.UseCase.Login` | Developer asks on the person's behalf | Person completes sign-in themselves through Claude Code | Interactive; uses the person's terminal |
| Sign out | `auth.UseCase.Logout` | Developer asks | Profile becomes not signed in | Already signed out is accepted quietly |
| Send a one-shot request | `query.UseCase.Ask` | Developer asks | Claude works through rounds and returns one final reply | Final reply only |
| Continue a conversation | `query.Request.SessionID` | Developer passes an earlier conversation reference | The new request joins that conversation | Unknown reference ⇒ "conversation not found" |
| Give up | context cancellation | Developer cancels or a time limit passes | Operation stops promptly, reported as given up | |

---

## 3. Ambiguities & Conflicts

| Ambiguous Term | Meaning in Context A | Meaning in Context B | Resolution |
| :--- | :--- | :--- | :--- |
| Client | The library's public entry point (`claude.Client`) | A wrapper around an external resource | `Client` only means the public entry point; wrappers of external resources are always `Proxy`. |
| Login / sign-in | Claude Code's own wording ("login") | Business wording ("sign-in") | Business documents say "sign-in"; code follows Claude Code and Go conventions (`Login`, `Logout`). |
| Token | Subscription pass | Developer account key | Never say "token" alone: "subscription pass" (OAuth token) vs "developer account key" (API key). |

---

## 4. External & Enum Mapping

| Category | Code Value / Key | Domain Label | Description |
| :--- | :--- | :--- | :--- |
| Account kind (Claude Code sign-in state) | `claude.ai` | Claude subscription | Signed in interactively with a Claude subscription |
| Account kind | `console` | Developer account | Signed in interactively with a pay-per-use developer account (value not yet observed) |
| Account kind | `oauth_token` | Subscription pass | Using a supplied subscription pass |
| Account kind | `api_key` | Developer account key | Using a supplied developer account key |
| Account kind | `none` | Not signed in | |
| Sign-in method (Claude Code sign-in) | `--claudeai` | Claude subscription | Default |
| Sign-in method | `--console` | Developer account | |
| Sign-in method | `--sso` | Company single sign-on | |
| Profile (Claude Code setting) | `CLAUDE_CONFIG_DIR` | Profile location | |
| Credential (Claude Code setting) | `CLAUDE_CODE_OAUTH_TOKEN` | Subscription pass | |
| Credential (Claude Code setting) | `ANTHROPIC_API_KEY` | Developer account key | |
| Request outcome (Claude Code result) | `success` | Finished | Final reply available |
| Request outcome | `error_max_turns` | Round limit reached | Stopped at the round limit before finishing |
| Request outcome | `is_error` with explanation | Claude reported a problem | e.g. "Not logged in · Please run /login", unavailable model |
