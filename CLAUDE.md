# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repo Is

`ai-session` is a session-based development workflow framework for AI assistants. It provides 16 `/session:*` commands that manage persistent state (feature description, implementation plan, open questions, log, reviews) across sessions via structured files — replacing reliance on chat history.

It supports both **Gemini CLI** and **Claude Code**. Claude commands (`.md` files) are the single source of truth; Gemini commands (`.toml` files) are generated from them via `scripts/gen_gemini.sh`, and headless pipeline variants are generated to `headless/session/` via `scripts/gen_headless.sh`. Commands are symlinked into each tool's native commands directory via `setup.sh`.

## Setup

```bash
chmod +x setup.sh
./setup.sh          # Adds AI_SESSION_HOME to .zshrc, creates symlinks
source ~/.zshrc
```

`setup.sh` is idempotent. Re-run it after adding a new command group — new subdirectories under `gemini/` or `claude/` are picked up automatically.

## Prerequisites

- `yq` v4+ (`brew install yq`) — must be [mikefarah/yq](https://github.com/mikefarah/yq), not the Python-based one
- `Node.js` and `uv` (for MCP servers)
- MCP servers configured in `~/.gemini/settings.json` or `~/.claude/settings.json`: `shortcut`, `notion`, `git`, `github`

## Repo Structure

```
spec/session/     # LLM-agnostic documentation, schemas, and example feature document
claude/session/          # Claude Code commands (*.md) — single source of truth for all prompts
headless/session/        # LLM-agnostic headless pipeline variants — generated via scripts/gen_headless.sh
gemini/session/   # Gemini CLI commands (*.toml) — generated from claude/session/ via scripts/gen_gemini.sh
scripts/          # Shared shell scripts referenced via $AI_SESSION_HOME/scripts/
```

## Architecture & Key Conventions

### Feature Directory Structure

Each feature gets a directory, stored by default in `~/.ai-session/features/<org>/<repo>/<feature-id>/` (centralized). If `.features/<feature-id>/` already exists in the target project's CWD, that path is used instead (backward compatibility).

```
~/.ai-session/features/talon-one/ai-sessions/sc-12345/
  description.md    # User story and acceptance criteria (Markdown, unstructured)
  architecture.md   # Optional: implementation strategy, pattern refs, constraints, slice hints
  plan.yml          # Execution plan: slices (id, description, status, depends_on) containing tasks (id, task, status)
  questions.yml     # Open questions: id, question, status, answer
  log.md            # Append-only progress log
  review.yml        # Code review findings: id, file, line, feedback, status
  pr.md             # Pull request link and description
```

YAML files are always modified via `yq` — never direct text replacement.

### Session Context Pattern

Commands follow producer/consumer roles to minimize token usage:
- **Producers** (`/session:start`, `/session:new`, `/session:define`): Read files from disk and output a `### ✨ Session Context Loaded for <feature-id>` block with `description.md` + `AGENTS.md` content.
- **Consumers** (`/session:plan`, `/session:review`, etc.): Read the context block from chat history — do NOT re-read files from disk.
- **Updaters** (`/session:pr`): After modifying a context file, output a new updated context block.

### Two Command Patterns

- **LLM Orchestrator**: Agent orchestrates helper scripts directly for interactive workflows (e.g., `/session:define`).
- **Sub-agent Pattern**: Agent delegates focused one-off tasks to an isolated sub-agent to keep the main session clean (e.g., `/session:review`, `/session:pr`).

### Project-Wide Context (`AGENTS.md`)

`AGENTS.md` in each **target project** (not this repo) stores architectural patterns, conventions, and learnings. `/session:end` updates it; `/session:start` reads it as part of the context block.

### Go CLI (`ai-session`) — Deterministic File Operations

The `ai-session` binary handles all structured file I/O so prompts never need raw `yq`, `sed`, or shell scripts for data mutations. Key subcommands:

```bash
ai-session serve [--port 1004]           # start read-only dashboard at http://localhost:1004
ai-session load-context sc-1234          # outputs feature dir files as XML blocks (replaces scripts/load_context_files.sh)
ai-session create-feature sc-1234        # scaffolds feature dir with placeholder files
ai-session resolve-feature-dir sc-1234  # prints the resolved feature dir path
ai-session append-log sc-1234 "msg"     # appends timestamped entry to log.md
ai-session update-task sc-1234 task-id --status done
ai-session update-slice sc-1234 slice-id --status in-progress
ai-session plan-list sc-1234            # lists slices (with --slice <id>: lists tasks)
ai-session plan-get sc-1234 --slice s --task t  # prints full task body
ai-session plan-write sc-1234           # validates + atomically writes plan.yml from stdin
ai-session plan-enrich-task sc-1234 --slice s --task t  # updates task body (stdin), injection guard
ai-session plan-split-task sc-1234 --slice s --task t   # replaces todo task with N subtasks (stdin YAML)
```

`scripts/load_context_files.sh` is **deprecated** — use `ai-session load-context` instead.

## Testing Commands

Commands are tested by delegating to a sub-agent using `spec/session/example-feature-document/` as the active feature context. Reference `spec/session/command_details.md` for expected behavior per command. There is no automated test runner — testing is manual via sub-agent delegation.

## Adding or Modifying Commands

- Edit/add `.md` files in `claude/session/` — this is the single source of truth.
- Run `scripts/gen_gemini.sh` to regenerate Gemini `.toml` files — only changed `.md` files are processed (checksum-based). Use `--force` to regenerate all.
- Run `scripts/gen_headless.sh` to regenerate headless pipeline variants in `headless/session/` — same checksum-based logic. Use `--force` to regenerate all. Commands in the deny list (`define`, `start`, `end`, `get-familiar`, `log-research`, `migration`, `plan`, `checkpoint`) are skipped; `plan.md` is hand-written.
- **Shared scripts**: add to `scripts/`, reference via `$AI_SESSION_HOME/scripts/<name>.sh`
- Do not edit `gemini/session/*.toml` directly — changes will be overwritten by the generator.
- `gemini/session/.checksums` is generated by the script and should be committed alongside `.toml` files.
