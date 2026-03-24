#!/usr/bin/env bash
# setup.sh — Idempotent setup for ai-session
# Run once after cloning, and again if you ever move the repo.
# Any subdirectory added to gemini/ or claude/ is automatically symlinked.
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ZSHRC="$HOME/.zshrc"
ZSHENV="$HOME/.zshenv"

# sed -i behaves differently on macOS vs Linux
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed_inplace() { sed -i '' "$@"; }
else
    sed_inplace() { sed -i "$@"; }
fi

echo "╔══════════════════════════════╗"
echo "║      ai-session  setup       ║"
echo "╚══════════════════════════════╝"
echo "Repo: $REPO_DIR"
echo ""

# ── 1. Set AI_SESSION_HOME for zsh ───────────────────────────────────────────
# We need to set the variable in both .zshenv (for Claude's non-interactive
# shells) and .zshrc (for Gemini's interactive shells).

EXPORT_LINE="export AI_SESSION_HOME=\"$REPO_DIR\""

# Add to .zshenv for Claude
if grep -q "AI_SESSION_HOME" "$ZSHENV" 2>/dev/null; then
    CURRENT_ZSHENV="$(grep "AI_SESSION_HOME" "$ZSHENV")"
    if [ "$CURRENT_ZSHENV" = "$EXPORT_LINE" ]; then
        echo "✓ AI_SESSION_HOME already correct in .zshenv (for Claude)"
    else
        sed_inplace "s|.*AI_SESSION_HOME.*|$EXPORT_LINE|" "$ZSHENV"
        echo "✓ Updated AI_SESSION_HOME in .zshenv → $REPO_DIR (for Claude)"
    fi
else
    printf "\n# ai-session\n%s\n" "$EXPORT_LINE" >> "$ZSHENV"
    echo "✓ Added AI_SESSION_HOME to .zshenv → $REPO_DIR (for Claude)"
fi

# Add to .zshrc for Gemini
if grep -q "AI_SESSION_HOME" "$ZSHRC" 2>/dev/null; then
    CURRENT_ZSHRC="$(grep "AI_SESSION_HOME" "$ZSHRC")"
    if [ "$CURRENT_ZSHRC" = "$EXPORT_LINE" ]; then
        echo "✓ AI_SESSION_HOME already correct in .zshrc (for Gemini)"
    else
        sed_inplace "s|.*AI_SESSION_HOME.*|$EXPORT_LINE|" "$ZSHRC"
        echo "✓ Updated AI_SESSION_HOME in .zshrc → $REPO_DIR (for Gemini)"
    fi
else
    printf "\n# ai-session\n%s\n" "$EXPORT_LINE" >> "$ZSHRC"
    echo "✓ Added AI_SESSION_HOME to .zshrc → $REPO_DIR (for Gemini)"
fi
echo ""


# ── 2. Create symlinks ────────────────────────────────────────────────────────
# ~/.gemini/commands/ and ~/.claude/commands/ stay as real directories so you
# can add personal commands freely alongside the repo-managed ones.
# Each subdirectory of gemini/ and claude/ gets its own symlink automatically —
# no script changes needed when adding new command groups.

create_symlink() {
    local target="$1"
    local link="$2"

    if [ -L "$link" ]; then
        if [ "$(readlink "$link")" = "$target" ]; then
            echo "  ✓ $link"
            return
        else
            rm "$link"
        fi
    elif [ -e "$link" ]; then
        local backup="${link}.backup.$(date +%Y%m%d%H%M%S)"
        mv "$link" "$backup"
        echo "  ⚠ Backed up existing path: $backup"
    fi

    mkdir -p "$(dirname "$link")"
    ln -s "$target" "$link"
    echo "  ✓ $link → $target"
}

link_subdirs() {
    local source_dir="$1"
    local target_dir="$2"

    mkdir -p "$target_dir"
    for dir in "$source_dir"/*/; do
        [ -d "$dir" ] || continue
        local dirname
        dirname="$(basename "$dir")"
        create_symlink "$dir" "$target_dir/$dirname"
    done
}

HAS_GEMINI=false
HAS_CLAUDE=false
command -v gemini &>/dev/null && HAS_GEMINI=true
command -v claude  &>/dev/null && HAS_CLAUDE=true

if [ "$HAS_GEMINI" = false ] && [ "$HAS_CLAUDE" = false ]; then
    echo "⚠ Neither 'gemini' nor 'claude' found in PATH."
    echo "  Install at least one tool and re-run this script."
    exit 1
fi

if [ "$HAS_GEMINI" = true ]; then
    echo "Gemini commands:"
    link_subdirs "$REPO_DIR/gemini" "$HOME/.gemini/commands"
else
    echo "○ Gemini CLI not found — skipping"
fi

echo ""

if [ "$HAS_CLAUDE" = true ]; then
    echo "Claude commands:"
    link_subdirs "$REPO_DIR/claude" "$HOME/.claude/commands"
else
    echo "○ Claude Code not found — skipping"
fi

# ── 3. Done ───────────────────────────────────────────────────────────────────

echo ""
echo "Done. To apply changes, run:"
echo "source ~/.zshenv && source ~/.zshrc"
echo "You may also need to restart the Gemini and/or Claude CLI for changes to take effect."
