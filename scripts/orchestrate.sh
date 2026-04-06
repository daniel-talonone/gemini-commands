#!/usr/bin/env bash
# orchestrate.sh — runs individual session pipeline steps for a story.
#
# Usage:
#   orchestrate.sh <story-id> --plan
#   orchestrate.sh <story-id> --implement
#   orchestrate.sh <story-id> --review
#   orchestrate.sh <story-id> --pr
#   orchestrate.sh --status <story-id>

# Robustly determine AI_SESSION_HOME based on the script's location.
if [ -z "${AI_SESSION_HOME:-}" ]; then
  AI_SESSION_HOME="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  export AI_SESSION_HOME
fi

set -euxo pipefail # Added 'x' for debug tracing

usage() {
  echo "Usage:"
  echo "  orchestrate.sh <story-id> --plan"
  echo "  orchestrate.sh <story-id> --implement"
  echo "  orchestrate.sh <story-id> --review"
  echo "  orchestrate.sh <story-id> --pr"
  echo "  orchestrate.sh --status <story-id>"
  exit 1
}

show_status() {
  local story_id="$1"
  local feature_dir
  feature_dir="$($AI_SESSION_HOME/go-session/bin/ai-session resolve-feature-dir "$story_id")" # Updated to use Go CLI

  if [ ! -d "$feature_dir" ]; then
    echo "Error: feature directory not found: $feature_dir" >&2
    exit 1
  fi

  if [ ! -f "$feature_dir/status.yaml" ]; then
    echo "Error: status.yaml not found in $feature_dir" >&2
    exit 1
  fi

  cat "$feature_dir/status.yaml"
}

run_headless() {
  local cmd="$1"
  local story_id="$2"
  # Corrected path to generated headless prompts
  local prompt_file="$AI_SESSION_HOME/headless/session/$cmd.md" 

  if [ ! -f "$prompt_file" ]; then
    echo "Error: Headless prompt not found at $prompt_file. Please run 'scripts/gen_headless.sh' to generate these prompts." >&2
    exit 1
  fi
  
  local prompt
  prompt="$(sed "s/{{args}}/$story_id/g" "$prompt_file")"
  
  if [ -z "$prompt" ]; then
      echo "Error: Prompt is empty after processing $prompt_file" >&2
      exit 1
  fi

  gemini --yolo -p "$prompt"
}

write_status() {
  local feature_dir="$1"
  local step="$2"
  local repo_slug="$3"
  local branch="$4"
  local now
  now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  local started_at="$now"
  if [ -f "$feature_dir/status.yaml" ]; then
    local existing
    existing="$(yq '.started_at' "$feature_dir/status.yaml" 2>/dev/null || true)"
    [ -n "$existing" ] && [ "$existing" != "null" ] && started_at="$existing"
  fi

  mkdir -p "$feature_dir"
  # Use a single-line printf to avoid shell parsing issues with line continuations.
  printf "mode: auto
repo: %s
branch: %s
pid: %d
pipeline_step: %s
started_at: %s
updated_at: %s
" "$repo_slug" "$branch" "$$" "$step" "$started_at" "$now" > "$feature_dir/status.yaml"
}

check_preconditions() {
  local flag="$1"
  local feature_dir="$2"

  if [ ! -d "$feature_dir" ]; then
    echo "Error: feature directory not found: $feature_dir. Run /session:new or /session:define first." >&2
    exit 1
  fi

  case "$flag" in
    --implement)
      if [ ! -f "$feature_dir/plan.yml" ]; then
        echo "Error: plan.yml not found in $feature_dir. Run orchestrate.sh $STORY_ID --plan first." >&2
        exit 1
      fi
      ;;
    --review)
      if [ ! -f "$feature_dir/plan.yml" ]; then
        echo "Error: plan.yml not found in $feature_dir. Run orchestrate.sh $STORY_ID --plan first." >&2
        exit 1
      fi
      if ! grep -q "IMPLEMENT COMPLETE" "$feature_dir/log.md" 2>/dev/null; then
        echo "Error: implement has not completed successfully. Run orchestrate.sh $STORY_ID --implement first." >&2
        exit 1
      fi
      ;;
    --pr)
      if [ ! -f "$feature_dir/plan.yml" ]; then
        echo "Error: plan.yml not found in $feature_dir. Run orchestrate.sh $STORY_ID --plan first." >&2
        exit 1
      fi
      if [ ! -f "$feature_dir/review.yml" ]; then
        echo "Error: review.yml not found in $feature_dir. Run orchestrate.sh $STORY_ID --review first." >&2
        exit 1
      fi
      ;;
  esac
}

# --- Arg parsing ---

if [ "${1:-}" = "--status" ]; then
  [ -z "${2:-}" ] && usage
  show_status "$2"
  exit 0
fi

STORY_ID="${1:-}"
FLAG="${2:-}"
[ -z "$STORY_ID" ] || [ -z "$FLAG" ] && usage

case "$FLAG" in
  --plan|--implement|--review|--pr) ;;
  *) usage ;;
esac

# --- Repo context ---

git remote get-url origin &>/dev/null || {
  echo "Error: not inside a git repository or no 'origin' remote." >&2
  exit 1
}

REMOTE_URL="$(git remote get-url origin)"
REMOTE_URL="${REMOTE_URL%.git}"
if [[ "$REMOTE_URL" == git@* ]]; then
  REPO_SLUG="${REMOTE_URL#*:}"
else
  REPO_SLUG="${REMOTE_URL#*://*/}"
fi

BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "$STORY_ID")"
FEATURE_DIR="$($AI_SESSION_HOME/go-session/bin/ai-session resolve-feature-dir "$STORY_ID")" # Updated to use Go CLI

check_preconditions "$FLAG" "$FEATURE_DIR"

# --- Step dispatch ---

case "$FLAG" in
  --plan)
    echo "==> [orchestrate] Running /session:plan for $STORY_ID"
    write_status "$FEATURE_DIR" "plan" "$REPO_SLUG" "$BRANCH"
    run_headless plan "$STORY_ID"
    write_status "$FEATURE_DIR" "plan-done" "$REPO_SLUG" "$BRANCH"
    ;;

  --implement)
    echo "==> [orchestrate] Running /session:implement for $STORY_ID"
    write_status "$FEATURE_DIR" "implement" "$REPO_SLUG" "$BRANCH"
    run_headless implement "$STORY_ID"
    write_status "$FEATURE_DIR" "implement-done" "$REPO_SLUG" "$BRANCH"
    ;;

  --review)
    echo "==> [orchestrate] Running /session:review for $STORY_ID"
    write_status "$FEATURE_DIR" "review" "$REPO_SLUG" "$BRANCH"
    run_headless review "$STORY_ID"
    write_status "$FEATURE_DIR" "review-done" "$REPO_SLUG" "$BRANCH"
    ;;

  --pr)
    echo "==> [orchestrate] Running /session:pr for $STORY_ID"
    write_status "$FEATURE_DIR" "pr" "$REPO_SLUG" "$BRANCH"
    run_headless pr "$STORY_ID"
    write_status "$FEATURE_DIR" "pr-done" "$REPO_SLUG" "$BRANCH"
    ;;
esac
