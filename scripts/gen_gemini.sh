#!/usr/bin/env bash
# gen_gemini.sh — Generate gemini/session/*.toml from claude/session/*.md
#
# Run this whenever a Claude command is added or modified.
# The gemini/session/ directory is fully generated — do not edit .toml files directly.
#
# Each prompt body is adapted from Claude to Gemini conventions by passing it
# through `gemini -p` with scripts/gemini_adapter_prompt.md as the instruction.
#
# Usage: ./scripts/gen_gemini.sh

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLAUDE_DIR="$REPO_DIR/claude/session"
GEMINI_DIR="$REPO_DIR/gemini/session"
ADAPTER_PROMPT_FILE="$REPO_DIR/scripts/gemini_adapter_prompt.md"

if ! command -v gemini &>/dev/null; then
    echo "Error: 'gemini' CLI not found in PATH. Install Gemini CLI to use this script." >&2
    exit 1
fi

if [ ! -f "$ADAPTER_PROMPT_FILE" ]; then
    echo "Error: Adapter prompt not found at $ADAPTER_PROMPT_FILE" >&2
    exit 1
fi

mkdir -p "$GEMINI_DIR"

adapter_prompt="$(cat "$ADAPTER_PROMPT_FILE")"

count=0
for md_file in "$CLAUDE_DIR"/*.md; do
    [ -f "$md_file" ] || continue

    name="$(basename "$md_file" .md)"
    toml_file="$GEMINI_DIR/$name.toml"

    # Extract description value from YAML frontmatter
    description="$(awk '
        BEGIN { f=0 }
        /^---/ { f++; next }
        f==1 && /^description:/ { sub(/^description: */, ""); print; exit }
    ' "$md_file")"

    if [ -z "$description" ]; then
        echo "  ⚠ Skipping $name.md — no description found in frontmatter"
        continue
    fi

    # Extract body — everything after the closing --- of the frontmatter
    body="$(awk 'BEGIN{f=0} /^---/{f++; next} f>=2{print}' "$md_file")"

    # Adapt the prompt body from Claude conventions to Gemini conventions via LLM.
    # gemini -p sets the instruction prompt; stdin is appended to it as content.
    # stderr is suppressed to hide Gemini CLI startup noise (MCP loading, keychain, etc.)
    printf "  Adapting %s.md..." "$name"
    adapted_body="$(printf '%s' "$body" | gemini -p "$adapter_prompt" 2>/dev/null)"
    printf " ✓\n"

    # Escape description for a TOML double-quoted string (backslash, then double-quote)
    description_escaped="$(printf '%s' "$description" | sed 's/\\/\\\\/g; s/"/\\"/g')"

    # Write TOML — triple-quoted strings are used for the prompt so no escaping is needed.
    # Leading newline after """ is trimmed by the TOML spec.
    {
        printf '# Generated from claude/session/%s.md — do not edit directly.\n' "$name"
        printf '# Run scripts/gen_gemini.sh to regenerate.\n'
        printf 'description = "%s"\n' "$description_escaped"
        printf 'prompt = """\n'
        printf '%s\n' "$adapted_body"
        printf '"""\n'
    } > "$toml_file"

    count=$((count + 1))
done

echo ""
echo "Generated $count Gemini command(s) in gemini/session/."
