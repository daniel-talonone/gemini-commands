#!/usr/bin/env bash
set -euo pipefail

if [ -z "${1:-}" ]; then
  echo "Usage: resolve_feature_dir.sh <feature-id|path>" >&2
  exit 1
fi

ARG="$1"

# Explicit path: contains / or starts with . or ~
if [[ "$ARG" == /* || "$ARG" == .* || "$ARG" == ~* || "$ARG" == */* ]]; then
  echo "$ARG"
  exit 0
fi

FEATURE_ID="$ARG"

# Backward compat: use local .features/ if it already exists
LOCAL_DIR=".features/$FEATURE_ID"
if [ -d "$LOCAL_DIR" ]; then
  echo "$LOCAL_DIR"
  exit 0
fi

# Derive centralized path from git remote
REMOTE_URL="$(git remote get-url origin 2>/dev/null)" || {
  echo "Error: not inside a git repository or no 'origin' remote." >&2
  exit 1
}

# Extracts 'org/repo' from both 'https://github.com/org/repo.git' and 'git@github.com:org/repo.git'
REMOTE_URL="${REMOTE_URL%.git}"
if [[ "$REMOTE_URL" == git@* ]]; then
  ORG_REPO="${REMOTE_URL#*:}"
else
  ORG_REPO="${REMOTE_URL#*://*/}"
fi

echo "$HOME/.ai-session/features/$ORG_REPO/$FEATURE_ID"
