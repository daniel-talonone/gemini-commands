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

# --- File List ---
# Defines the set of standard context files to be loaded.
FILES_TO_LOAD=(
  "description.md"
  "plan.yml"
  "questions.yml"
  "review.yml"
  "devops-review.yml"
  "log.md"
  "pr.md"
)

# --- Read and Print Files ---
# Loop through the defined files, check for existence, and print with a delimiter.
for file in "${FILES_TO_LOAD[@]}"; do
  FILE_PATH="$FEATURE_DIR/$file"
  if [ -f "$FILE_PATH" ]; then
    echo "--- FILE: $file ---"
    cat "$FILE_PATH"
    echo "" # Add a newline for better separation
  fi
done

# Also load the project context file (AGENTS.md takes precedence as the LLM-agnostic standard,
# falling back to GEMINI.md for backward compatibility).
# Search order: project root (CWD), then the parent of the feature directory (e.g. .vscode/).
AGENTS_FILE=""
AGENTS_LABEL=""
for candidate_dir in "." "$(dirname "$FEATURE_DIR")"; do
    if [ -f "$candidate_dir/AGENTS.md" ]; then
        AGENTS_FILE="$candidate_dir/AGENTS.md"
        AGENTS_LABEL="AGENTS.md"
        break
    elif [ -f "$candidate_dir/GEMINI.md" ]; then
        AGENTS_FILE="$candidate_dir/GEMINI.md"
        AGENTS_LABEL="GEMINI.md"
        break
    fi
done

if [ -n "$AGENTS_FILE" ]; then
    echo "--- FILE: $AGENTS_LABEL ---"
    cat "$AGENTS_FILE"
    echo ""
fi
