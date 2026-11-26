# Rules

Implementation rules for go-claude: architecture, naming, style, testing. They contain no
product knowledge — the domain vocabulary lives in `.sdd/UL-MAP.md`, user-facing
documentation in `README.md`, and the contribution workflow in `CONTRIBUTING.md`.

| File | Covers |
| :--- | :--- |
| [architecture.md](architecture.md) | Library-by-feature layout, dependency direction, composition root, public API contract, when to use interfaces, sealed unions, behavior placement |
| [naming.md](naming.md) | Role suffixes, no package-name stutter, `Proxy` for external resources, capability-named unexported interfaces, file names |
| [code-style.md](code-style.md) | No `any`, declare-and-assign, no helper grab-bags, error wrapping, secret redaction, context cancellation |
| [testing.md](testing.md) | Business-behavior tests, testify mocks only at the proxy, no fakes, file placement, integration tests against the real CLI |

These rules derive from a language-agnostic rule set adapted for a Go library. When a
rule collides with reality, change the rule in a commit — never let code and rules drift.
