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

# Also load the global project GEMINI.md, which is expected by session commands.
if [ -f "GEMINI.md" ]; then
    echo "--- FILE: GEMINI.md ---"
    cat "GEMINI.md"
    echo ""
fi
