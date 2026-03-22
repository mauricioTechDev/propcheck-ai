# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
make build        # Build binary to bin/propcheck-ai (injects version via ldflags)
make test         # Run all tests: go test ./... -v
make test-short   # Run tests without I/O: go test ./... -short
make test-race    # Run tests with race detector: go test -race ./...
make lint         # Lint with golangci-lint (errcheck, staticcheck, gocritic, revive, etc.)
make ci           # Run full CI suite locally: lint + test-race + build
make coverage     # Run tests with race detector and print per-function coverage
make install      # Copy binary to ~/go/bin/
make clean        # Remove bin/ and coverage.out
```

Run a single package's tests:
```bash
go test ./internal/phase/ -v -run TestNext
```

## Pre-Commit Checklist

Before committing or pushing code, ALWAYS run:

1. `gofmt -w .` — Auto-format all Go files
2. `make ci` — Run lint, tests (with race detector), and build

Do NOT commit or push if `make ci` fails. Fix issues first.

## Architecture

**propcheck-ai** is a Go CLI tool that enforces property-based testing (PBT) discipline for AI coding agents. It does NOT run tests itself — the AI agent runs tests using whatever framework the project uses. The tool provides a state machine, phase-appropriate guidance, and structured output.

### Package Layout

- `main.go` — Entry point, calls `cmd.Execute()`
- `cmd/` — Cobra CLI commands. Each command follows: load session → validate → mutate → save → output
- `internal/types/` — Core data structures: `Phase`, `Property`, `Session`, `Guidance`, `Event`
- `internal/session/` — Session persistence (read/write `.propcheck-ai.json` in working directory)
- `internal/phase/` — State machine: `Next()`, `NextWithResult()`, `NextInLoop()`, `ExpectedTestResult()`, `CanTransition()`
- `internal/guide/` — Generates phase-specific state output based on current phase
- `internal/reflection/` — PBT reflection questions and answer validation for the refine phase
- `internal/verify/` — Post-hoc PBT compliance analysis: checks property_picked events, test runs, shrink analysis; returns compliance score (0-100%)
- `internal/formatter/` — Formats output as text or JSON (`FormatGuidance`, `FormatStatus`, `FormatFullStatus`, `FormatResume`)

### Key Concepts

**PBT Phase State Machine:**

Standard flow (tests pass in VALIDATE):
```
PROPERTY → GENERATE → VALIDATE → REFINE → [PROPERTY (loop) | DONE]
```

Failure flow (tests fail in VALIDATE — counter-example found):
```
PROPERTY → GENERATE → VALIDATE → SHRINK → REFINE → [PROPERTY (loop) | DONE]
```

VALIDATE is the branching point: test result determines REFINE (pass) vs SHRINK (fail).

**Per-Property Tight Loop:**
1. `propcheck-ai property add "desc1" "desc2" ...` — Add properties to the backlog
2. `propcheck-ai property pick <id>` — Pick ONE property to work on
3. PROPERTY → GENERATE → VALIDATE → [SHRINK] → REFINE for that property
4. On `phase next` from REFINE: auto-completes the current property, loops back to PROPERTY if properties remain, or advances to DONE if the backlog is empty

**SHRINK Phase:** Only entered when VALIDATE finds a failure. Agent must record counter-example analysis via `propcheck-ai shrink analyze --answer "..."` before advancing.

**Refine Reflections:** During the refine phase, 7 PBT-specific reflection questions are loaded. Agents must answer all questions (min 5 words each) before advancing.

**Agent Mode:** `propcheck-ai init --agent` enables stricter enforcement: `phase set` is disabled entirely, `complete` requires `--force`.

**Compliance Verification:** `propcheck-ai verify` analyzes session history for PBT violations (missing property_picked, no test runs, missing shrink analysis). Returns a compliance score (0-100%) and exit code 1 on violations.

**Claude Code Hooks:** Two `PreToolUse` hooks in `.claude/hooks/`:
- `propcheck-guard.sh` — Blocks non-test file writes during PROPERTY and GENERATE phases
- `propcheck-commit-check.sh` — Blocks `git commit` when phase is not `done`

**Session File:** `.propcheck-ai.json` in the working directory stores phase, agent mode, properties, test command, last test result, current property ID, iteration count, shrink analysis, reflections, and event history.

**Output Format:** All commands support `--format json` for machine-readable output. Format auto-detects: JSON when piped, text when in a terminal.

### Dependencies

- `github.com/spf13/cobra` — CLI framework
- `golang.org/x/term` — Terminal detection for auto-format
- No external databases, APIs, or services
