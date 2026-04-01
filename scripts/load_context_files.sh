#!/bin/bash

# A script to load all context files from a feature directory
# and print them to stdout, separated by a delimiter.

set -e
set -u

# --- Argument Validation ---
if [ -z "$1" ]; then
  echo "Error: No feature directory path provided." >&2
  echo "Usage: $0 /path/to/feature_dir" >&2
  exit 1
fi

FEATURE_DIR=$1

if [ ! -d "$FEATURE_DIR" ]; then
  echo "Error: Directory not found at '$FEATURE_DIR'" >&2
  exit 1
fi

# --- Read and Print Files ---
# Load all .md and .yml/.yaml files in the feature directory, sorted alphabetically.
# Files starting with '_' (e.g. _SUMMARY.md) are excluded — they are generated artifacts.
while IFS= read -r FILE_PATH; do
  file="$(basename "$FILE_PATH")"
  echo "--- FILE: $file ---"
  cat "$FILE_PATH"
  echo ""
done < <(find "$FEATURE_DIR" -maxdepth 1 \( -name "*.md" -o -name "*.yml" -o -name "*.yaml" \) ! -name "_*" | sort)

# Also load the project context file (AGENTS.md takes precedence as the LLM-agnostic standard,
# falling back to GEMINI.md for backward compatibility).
# Search order: project root (CWD) only.
AGENTS_FILE=""
AGENTS_LABEL=""
if [ -f "./AGENTS.md" ]; then
    AGENTS_FILE="./AGENTS.md"
    AGENTS_LABEL="AGENTS.md"
elif [ -f "./GEMINI.md" ]; then
    AGENTS_FILE="./GEMINI.md"
    AGENTS_LABEL="GEMINI.md"
fi

if [ -n "$AGENTS_FILE" ]; then
    echo "--- FILE: $AGENTS_LABEL ---"
    cat "$AGENTS_FILE"
    echo ""
fi
