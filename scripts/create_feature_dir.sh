#!/usr/bin/env bash
# Usage: ./create_feature_dir.sh <full-feature-path>
# Example: ./create_feature_dir.sh ~/.features/org/repo/sc-12345
#
# Derives repo (org/repo), branch, and work_dir from git if available.
# Does NOT overwrite files that already exist.

if [ -z "$1" ]; then
  echo "Error: Full feature directory path is required." >&2
  exit 1
fi

DIR_PATH="$1"

# Derive repo slug and branch from git (best-effort, no error if unavailable)
REMOTE_URL=$(git remote get-url origin 2>/dev/null || true)
REPO_SLUG=""
if [ -n "$REMOTE_URL" ]; then
  REMOTE_URL="${REMOTE_URL%.git}"
  if [[ "$REMOTE_URL" == git@* ]]; then
    REPO_SLUG="${REMOTE_URL##*:}"
  else
    REPO_SLUG=$(echo "$REMOTE_URL" | sed 's|.*://[^/]*/||')
  fi
fi

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)
if [ "$BRANCH" = "HEAD" ]; then
  BRANCH=""
fi

WORK_DIR=$(git rev-parse --show-toplevel 2>/dev/null || true)

AI_SESSION_BIN="$(dirname "$0")/../go-session/bin/ai-session"

# create-feature creates the directory, all placeholder files, and status.yaml in one call.
# It is idempotent: existing files are never overwritten.
"$AI_SESSION_BIN" create-feature "$DIR_PATH" \
  --repo "$REPO_SLUG" \
  --branch "$BRANCH" \
  --work-dir "$WORK_DIR"

echo "Created and populated placeholder files in $DIR_PATH"
