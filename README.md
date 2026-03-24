> [!IMPORTANT]
> ## BREAKING CHANGE: Repo Location & Multi-Tool Support
>
> Previously, this repo was cloned directly into `~/.gemini/commands/` and only supported Gemini CLI.
>
> It now lives in its own neutral location (`~/.ai-session/`) and supports **both Gemini CLI and Claude Code**.
> A `setup.sh` script handles all wiring automatically.
>
> **If you cloned this into `~/.gemini/commands/` before, migrate with:**
> ```sh
> mv ~/.gemini/commands ~/.ai-session
> ~/.ai-session/setup.sh
> source ~/.zshrc
> ```

# ai-session

A session-based development workflow for AI assistants. Instead of relying on fragile
chat history, this system gives your AI structured files — a feature description,
implementation plan, open questions, and a running log — that persist across sessions
and make context explicit and reusable.

Compatible with **Gemini CLI**, **Claude Code**, or both simultaneously. Each tool gets
its own set of commands (`gemini/` and `claude/`) that implement the same workflow
concepts in that tool's native format.

## Prerequisites

> **Platform note:** these instructions are written for macOS. `setup.sh` works on
> Linux too, but the install commands below use Homebrew — substitute your distro's
> package manager (`apt`, `dnf`, etc.) as needed.

Before running setup, make sure you have:

- **[Gemini CLI](https://github.com/google-gemini/gemini-cli)** and/or **[Claude Code](https://github.com/anthropics/claude-code)** — install whichever tools you plan to use
- **[yq](https://github.com/mikefarah/yq) v4+** — used to update YAML state files (`brew install yq`)
- **[Node.js](https://nodejs.org/)** — required for `npx` (used by MCP servers)
- **[uv](https://docs.astral.sh/uv/)** — required for `uvx` (used by the Git MCP server)
- **git**

## Setup

```bash
git clone git@github.com:daniel-talonone/gemini-commands.git ~/.ai-session
chmod +x ~/.ai-session/setup.sh
~/.ai-session/setup.sh
source ~/.zshrc
```

**Gemini CLI users:** also install the required skills after setup:
```bash
gemini skills install ~/.ai-session/gemini/tdd-skill
gemini skills install ~/.ai-session/gemini/yq-skill
```

`setup.sh` does two things:
1. Adds `export AI_SESSION_HOME="$HOME/.ai-session"` to your `.zshrc`
2. Creates symlinks from each tool's commands directory into this repo

> **Not using zsh?** Manually add `export AI_SESSION_HOME="$HOME/.ai-session"` to
> your shell's config file (`.bashrc`, `.bash_profile`, `config.fish`, etc.) before
> running any commands.

## Structure

```
.
├── spec/          # LLM-agnostic: documentation, schemas, examples
│   └── session/
├── gemini/        # Gemini CLI implementation (*.toml)
│   └── session/
├── claude/        # Claude Code implementation (*.md)
│   └── session/
├── scripts/       # Shared helper scripts used by both tools
├── roadmap/       # Project roadmap and reviews
├── AGENTS.md      # Project-wide AI context
├── setup.sh       # Idempotent setup script
└── README.md
```

## How it works

`setup.sh` symlinks each subdirectory of `gemini/` and `claude/` into the respective
tool's commands directory:

- `~/.gemini/commands/<group>/` → `~/.ai-session/gemini/<group>/`
- `~/.claude/commands/<group>/` → `~/.ai-session/claude/<group>/`

Each tool's `commands/` directory remains a real folder, so you can add personal
commands alongside the repo-managed ones without touching this repo.

Adding a new command group (e.g. `gemini/transaction/`) is automatically picked up
the next time `setup.sh` is run — no script changes needed.

All commands reference shared scripts via `$AI_SESSION_HOME/scripts/`.

## Session workflow

See [`spec/session/README.md`](spec/session/README.md) for full documentation of the
session commands, workflow lifecycle, and core concepts.
