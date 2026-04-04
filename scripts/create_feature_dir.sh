#!/usr/bin/env bash
# Usage: ./create_feature_dir.sh <full-feature-path>
# Example: ./create_feature_dir.sh ~/.ai-session/features/org/repo/sc-12345
#
# Derives repo (org/repo) and branch from git if available.
# Does NOT overwrite files that already exist.

if [ -z "$1" ]; then
  echo "Error: Full feature directory path is required." >&2
  exit 1
fi

DIR_PATH="$1"

mkdir -p "$DIR_PATH"

# Derive repo slug and branch from git (best-effort, no error if unavailable)
REMOTE_URL=$(git remote get-url origin 2>/dev/null || true)
REPO_SLUG=""
if [ -n "$REMOTE_URL" ]; then
  # Strip .git suffix
  REMOTE_URL="${REMOTE_URL%.git}"
  if [[ "$REMOTE_URL" == git@* ]]; then
    # SSH: git@github.com:org/repo → org/repo
    REPO_SLUG="${REMOTE_URL##*:}"
  else
    # HTTPS: https://github.com/org/repo → org/repo
    REPO_SLUG=$(echo "$REMOTE_URL" | sed 's|.*://[^/]*/||')
  fi
fi

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)
if [ "$BRANCH" = "HEAD" ]; then
  BRANCH=""
fi

WORK_DIR=$(git rev-parse --show-toplevel 2>/dev/null || true)

# Create files only if they don't already exist
create_if_missing() {
  local path="$1"
  local content="$2"
  if [ ! -f "$path" ]; then
    printf '%s' "$content" > "$path"
  fi
}

create_if_missing "$DIR_PATH/plan.yml" "[]\n"
create_if_missing "$DIR_PATH/questions.yml" "[]\n"
create_if_missing "$DIR_PATH/review.yml" "[]\n"
create_if_missing "$DIR_PATH/log.md" "# Work Log\n*(This section is intentionally left blank.)*\n"
create_if_missing "$DIR_PATH/pr.md" "# Pull Request\n*(This section is intentionally left blank.)*\n"

if [ ! -f "$DIR_PATH/status.yaml" ]; then
  REPO_VAL="${REPO_SLUG:-''}"
  BRANCH_VAL="${BRANCH:-''}"
  WORK_DIR_VAL="${WORK_DIR:-''}"
  NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  cat <<EOF > "$DIR_PATH/status.yaml"
mode: ''
repo: ${REPO_VAL}
branch: ${BRANCH_VAL}
work_dir: ${WORK_DIR_VAL}
pid: 0
pipeline_step: ''
started_at: '${NOW}'
updated_at: '${NOW}'
EOF
fi

echo "Created and populated placeholder files in $DIR_PATH"
