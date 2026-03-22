#!/usr/bin/env bash
set -euo pipefail

# propcheck-commit-check.sh — Claude Code PreToolUse hook
# Blocks git commit when PBT phase is not "done".

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // ""')

if [[ "$COMMAND" != *"git commit"* && "$COMMAND" != *"git -C"*"commit"* ]]; then
  exit 0
fi

SESSION_FILE=".propcheck-ai.json"
if [[ ! -f "$SESSION_FILE" ]]; then
  exit 0
fi

PHASE=$(jq -r '.phase // ""' "$SESSION_FILE")
if [[ "$PHASE" == "done" ]]; then
  exit 0
fi

echo "BLOCKED: PBT cycle not complete. Current phase is $PHASE. Complete all properties through PROPERTY-GENERATE-VALIDATE-SHRINK-REFINE before committing." >&2
exit 2
