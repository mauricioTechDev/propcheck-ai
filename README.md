# propcheck-ai

Property-based testing (PBT) guardrails for AI coding agents.

**propcheck-ai** is a CLI state machine that enforces PBT discipline. It does NOT run tests — the AI agent runs tests using whatever PBT framework the project uses (or manual random loops if no framework exists). The tool provides phase tracking, property management, and enforcement hooks.

## Quick Start

```bash
# Initialize a session
propcheck-ai init --test-cmd "go test ./..."

# Add properties to test
propcheck-ai property add "sort is idempotent: sort(sort(xs)) == sort(xs)" --category invariant
propcheck-ai property add "reverse is its own inverse: reverse(reverse(xs)) == xs" --category roundtrip

# Pick a property and start the loop
propcheck-ai property pick 1

# The PBT cycle
propcheck-ai phase next                        # PROPERTY → GENERATE
propcheck-ai phase next                        # GENERATE → VALIDATE
propcheck-ai test                              # Run tests
propcheck-ai phase next --test-result pass     # VALIDATE → REFINE
propcheck-ai refine reflect 1 --answer "..."   # Answer reflections
propcheck-ai phase next --test-result pass     # REFINE → next property or DONE
```

## Phase State Machine

```
Standard flow (tests pass):
PROPERTY → GENERATE → VALIDATE → REFINE → [PROPERTY (loop) | DONE]

Failure flow (counter-example found):
PROPERTY → GENERATE → VALIDATE → SHRINK → REFINE → [PROPERTY (loop) | DONE]
```

| Phase | What to do | File gating |
|-------|-----------|-------------|
| **PROPERTY** | Define the invariant to test | Test files only |
| **GENERATE** | Write input generators | Test files only |
| **VALIDATE** | Run with many random inputs | Any files |
| **SHRINK** | Analyze minimal counter-example | Any files |
| **REFINE** | Fix code, improve generators, answer reflections | Any files |
| **DONE** | All properties complete | — |

## When Tests Fail in VALIDATE

Finding a failure is a **success** in PBT — you discovered a bug. The CLI routes you through SHRINK:

```bash
propcheck-ai phase next --test-result fail     # VALIDATE → SHRINK
propcheck-ai shrink analyze --answer "The minimal failing input was [-3, 0, 5]. The comparator used > instead of >=."
propcheck-ai phase next                        # SHRINK → REFINE
```

## Language Agnostic

propcheck-ai works with **any language**. The CLI manages workflow discipline, not technology:

- **With a PBT library**: Use `hypothesis` (Python), `fast-check` (JS), `rapid` (Go), `proptest` (Rust), `jqwik` (Java)
- **Without a PBT library**: Write manual random testing loops — the PBT *discipline* still applies

## Property Categories

Organize properties by category:

```bash
propcheck-ai property add --category invariant "sort is idempotent"
propcheck-ai property add --category roundtrip "parse(format(x)) == x"
propcheck-ai property add --category equivalence "naive_sort(xs) == optimized_sort(xs)"
propcheck-ai property add --category metamorphic "sort(xs ++ ys) == sort(sort(xs) ++ sort(ys))"
```

## Agent Mode

For stricter enforcement with AI agents:

```bash
propcheck-ai init --agent --test-cmd "npm test"
```

Agent mode disables `phase set` and requires `--force` for `complete`.

## Compliance Verification

```bash
propcheck-ai verify
# PBT Compliance: PASS
# Score: 100% (3/3 properties compliant)
```

Checks: property_picked events, test runs during VALIDATE, shrink analysis when failures found, no phase_set bypass usage.

## Installation

### One-liner (macOS / Linux)

```bash
curl -sSL https://raw.githubusercontent.com/mauricioTechDev/propcheck-ai/main/install.sh | sh
```

Install to a custom directory:

```bash
curl -sSL https://raw.githubusercontent.com/mauricioTechDev/propcheck-ai/main/install.sh | INSTALL_DIR=~/.local/bin sh
```

Install a specific version:

```bash
curl -sSL https://raw.githubusercontent.com/mauricioTechDev/propcheck-ai/main/install.sh | VERSION=0.1.0 sh
```

### Download from GitHub Releases

Pre-built binaries for macOS, Linux, and Windows are available on the [Releases page](https://github.com/mauricioTechDev/propcheck-ai/releases).

### From source

```bash
go install github.com/mauricioTechDev/propcheck-ai@latest
```

### Build from repo

```bash
git clone https://github.com/mauricioTechDev/propcheck-ai.git
cd propcheck-ai
make install
```

## Claude Code Hooks

Two enforcement hooks are provided in `.claude/hooks/`:

- **propcheck-guard.sh** — Blocks non-test file writes during PROPERTY and GENERATE phases
- **propcheck-commit-check.sh** — Blocks `git commit` when phase is not `done`

Register them in `.claude/settings.local.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit",
        "hook": ".claude/hooks/propcheck-guard.sh"
      },
      {
        "matcher": "Bash",
        "hook": ".claude/hooks/propcheck-commit-check.sh"
      }
    ]
  }
}
```
