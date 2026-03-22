#!/usr/bin/env bash
set -euo pipefail

# propcheck-guard.sh — Claude Code PreToolUse hook
# Blocks non-test file writes during PROPERTY and GENERATE phases.

INPUT=$(cat)

TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // ""')
if [[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]]; then
  exit 0
fi

SESSION_FILE=".propcheck-ai.json"
if [[ ! -f "$SESSION_FILE" ]]; then
  exit 0
fi

PHASE=$(jq -r '.phase // ""' "$SESSION_FILE")
if [[ "$PHASE" != "property" && "$PHASE" != "generate" ]]; then
  exit 0
fi

FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // ""')
if [[ -z "$FILE_PATH" ]]; then
  exit 0
fi

BASENAME=$(basename "$FILE_PATH")
if [[ "$BASENAME" == *_test.* ]] || \
   [[ "$BASENAME" == *.test.* ]] || \
   [[ "$BASENAME" == *.spec.* ]] || \
   [[ "$FILE_PATH" == */test/* ]] || \
   [[ "$FILE_PATH" == */tests/* ]]; then
  exit 0
fi

echo "BLOCKED: During the PROPERTY/GENERATE phases, only test files may be written. Write your property tests and generators first, then advance to later phases to modify implementation code." >&2
exit 2
