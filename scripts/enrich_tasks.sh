#!/usr/bin/env bash
# enrich_tasks.sh — Enrich plan.yml task descriptions with FILE/FUNCTION/code context.
#
# Reads plan.yml from the feature directory, collects source files referenced in
# task descriptions, passes everything through an LLM adapter prompt, and writes
# back an enriched plan.yml atomically (temp file + mv to prevent corruption).
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
TEMP_FILE="$FEATURE_DIR/plan.yml.enriching"

if [ ! -f "$PLAN_FILE" ]; then
    echo "Error: plan.yml not found at '$PLAN_FILE'" >&2
    exit 1
fi

if ! command -v gemini &>/dev/null; then
    echo "Error: 'gemini' CLI not found in PATH." >&2
    exit 1
fi

# Clean up temp file on exit (covers failures mid-run)
trap 'rm -f "$TEMP_FILE"' EXIT

enricher_prompt="$(cat "$PROMPT_FILE")"

# Build input: plan.yml + any source files mentioned in task descriptions.
# We grep for patterns like "FILE: src/..." or "src/..." in the task text and
# collect those files from the current working directory (the target project).
input="$(printf '--- plan.yml ---\n%s\n' "$(cat "$PLAN_FILE")")"

while IFS= read -r file_path; do
    if [ -f "$file_path" ]; then
        input="$(printf '%s\n\n--- FILE: %s ---\n%s' "$input" "$file_path" "$(cat "$file_path")")"
    fi
done < <(grep -oE '(src|lib|app|test|tests|spec|packages)/[^[:space:]]+\.[a-z]+' "$PLAN_FILE" | sort -u)

# Run enrichment via LLM. Stderr suppressed to hide Gemini CLI startup noise.
printf '%s' "$input" | gemini -p "$enricher_prompt" 2>/dev/null > "$TEMP_FILE"

# Validate the output is non-empty before replacing
if [ ! -s "$TEMP_FILE" ]; then
    echo "Error: LLM returned empty output — plan.yml not modified." >&2
    exit 1
fi

# Atomic replace
mv "$TEMP_FILE" "$PLAN_FILE"

# Log completion
"$REPO_DIR/scripts/append_to_log.sh" "$FEATURE_DIR/log.md" \
    "**Background task enrichment complete.** \`plan.yml\` has been enriched with FILE/FUNCTION/code context. Review it before starting implementation."
