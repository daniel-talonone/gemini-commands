#!/usr/bin/env bash
# enrich_tasks.sh — Enrich plan.yml task descriptions with FILE/FUNCTION/code context.
#
# Iterates over todo tasks one at a time using the ai-session CLI.
# The LLM decides whether to ENRICH (clarify), SPLIT (into subtasks), or SKIP (already detailed).
# No full-file overwrite, no schema corruption possible.
#
# Intended to run as a detached background process after /session:plan:
#   nohup $AI_SESSION_HOME/scripts/enrich_tasks.sh ".features/sc-1234" >> ".features/sc-1234/log.md" 2>&1 &
#
# Usage: scripts/enrich_tasks.sh <feature-dir>

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROMPT_FILE="$REPO_DIR/scripts/enrich_tasks_prompt.md"

if [ -z "${1:-}" ]; then
    echo "Error: No feature directory path provided." >&2
    echo "Usage: $0 <feature-dir>" >&2
    exit 1
fi

FEATURE_DIR="$1"
PLAN_FILE="$FEATURE_DIR/plan.yml"

if [ ! -f "$PLAN_FILE" ]; then
    echo "Error: plan.yml not found at '$PLAN_FILE'" >&2
    exit 1
fi

if ! command -v gemini &>/dev/null; then
    echo "Error: 'gemini' CLI not found in PATH." >&2
    exit 1
fi

if ! command -v ai-session &>/dev/null; then
    echo "Error: 'ai-session' CLI not found in PATH. Run setup.sh to add go-session/bin/ to PATH." >&2
    exit 1
fi

enricher_prompt="$(cat "$PROMPT_FILE")"
enriched_count=0
skipped_count=0

# Iterate slices
while IFS= read -r slice_line; do
    # plan-list output: "<id>  [<status>]" or "  <description>" — skip description lines
    [[ "$slice_line" =~ ^[[:space:]] ]] && continue
    [[ -z "$slice_line" ]] && continue
    slice_id="$(echo "$slice_line" | awk '{print $1}')"
    [ -z "$slice_id" ] && continue

    # Iterate tasks in this slice
    while IFS= read -r task_line; do
        [[ -z "$task_line" ]] && continue
        task_id="$(echo "$task_line" | awk '{print $1}')"
        task_status="$(echo "$task_line" | grep -oE '\[.*\]' | tr -d '[]')"
        [ -z "$task_id" ] && continue

        # Only enrich todo tasks
        if [ "$task_status" != "todo" ]; then
            skipped_count=$((skipped_count + 1))
            continue
        fi

        # Get the current task body
        task_body="$(ai-session plan-get "$FEATURE_DIR" --slice "$slice_id" --task "$task_id" 2>/dev/null || true)"

        # Build slice context: sibling task IDs + first line of their body
        slice_context="$(ai-session plan-list "$FEATURE_DIR" --slice "$slice_id" 2>/dev/null | \
            grep -v '^[[:space:]]' | awk '{print $1}' | while IFS= read -r sid; do
                body="$(ai-session plan-get "$FEATURE_DIR" --slice "$slice_id" --task "$sid" 2>/dev/null | head -1 || true)"
                printf '%s: %s\n' "$sid" "$body"
            done)"

        # Build prompt input with slice context prepended
        input="--- slice context: $slice_id ---
$slice_context
---
--- task: $slice_id/$task_id ---
$task_body"

        # Append referenced source files
        while IFS= read -r file_path; do
            if [ -f "$file_path" ]; then
                input="$input

--- FILE: $file_path ---
$(cat "$file_path")"
            fi
        done < <(echo "$task_body" | grep -oE '(src|lib|app|test|tests|spec|packages)/[^[:space:]]+\.[a-z]+' | sort -u)

        # Call LLM
        new_body="$(printf '%s\n' "$input" | gemini -p "$enricher_prompt" 2>/dev/null | sed '/^```/d')"

        if [ -z "$new_body" ]; then
            echo "Warning: LLM returned empty output for task $slice_id/$task_id — skipping." >&2
            skipped_count=$((skipped_count + 1))
            continue
        fi

        # Detect ENRICH: / SPLIT: prefix and route to the correct command
        first_line="$(printf '%s' "$new_body" | head -1)"
        rest="$(printf '%s' "$new_body" | tail -n +2)"

        if [ "$first_line" = "ENRICH:" ]; then
            if printf '%s' "$rest" | ai-session plan-enrich-task "$FEATURE_DIR" \
                    --slice "$slice_id" --task "$task_id" 2>/tmp/enrich_err; then
                enriched_count=$((enriched_count + 1))
            else
                err="$(cat /tmp/enrich_err)"
                echo "Warning: plan-enrich-task failed for $slice_id/$task_id: $err" >&2
                skipped_count=$((skipped_count + 1))
            fi
        elif [ "$first_line" = "SPLIT:" ]; then
            if printf '%s' "$rest" | ai-session plan-split-task "$FEATURE_DIR" \
                    --slice "$slice_id" --task "$task_id" 2>/tmp/enrich_err; then
                enriched_count=$((enriched_count + 1))
            else
                err="$(cat /tmp/enrich_err)"
                echo "Warning: plan-split-task failed for $slice_id/$task_id: $err" >&2
                skipped_count=$((skipped_count + 1))
            fi
        elif [ "$first_line" = "SKIP:" ]; then
            skipped_count=$((skipped_count + 1))
        else
            echo "Warning: LLM output for $slice_id/$task_id missing ENRICH:/SPLIT:/SKIP: prefix — skipping." >&2
            skipped_count=$((skipped_count + 1))
        fi

    done < <(ai-session plan-list "$FEATURE_DIR" --slice "$slice_id" 2>/dev/null || true)

done < <(ai-session plan-list "$FEATURE_DIR" 2>/dev/null || true)

rm -f /tmp/enrich_err

"$REPO_DIR/scripts/append_to_log.sh" "$FEATURE_DIR/log.md" \
    "**Background task enrichment complete.** $enriched_count tasks enriched, $skipped_count skipped. Review \`plan.yml\` before starting implementation."
