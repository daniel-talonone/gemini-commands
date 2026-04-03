# Project Overview

This project is a suite of AI assistant commands that implement a structured,
session-based workflow for software development, compatible with both Gemini CLI and
Claude Code. The commands orchestrate a lifecycle for feature development, from
creation and planning to review and pull request generation.

The core of the workflow revolves around a "feature directory" (stored by default in
`~/.ai-session/features/<org>/<repo>/<feature-id>/`) which acts as a single source
of truth for a given task, and an `AGENTS.md` file in each project which provides
global context. See `spec/session/example-feature-document/` for a complete example
of a feature directory.

## Repo Structure

```
claude/session/          ← Claude Code commands (*.md) — single source of truth for all prompts
headless/session/        ← LLM-agnostic headless pipeline variants — generated via scripts/gen_headless.sh
gemini/session/   ← Gemini CLI commands (*.toml) — generated via scripts/gen_gemini.sh
spec/session/     ← LLM-agnostic documentation, schemas, and examples
scripts/          ← Shared shell scripts used by both tools
cmd/ai-session/   ← Go CLI for deterministic file operations
```

Feature directories contain:
- `description.md` — requirements (what)
- `architecture.md` — optional: implementation strategy, pattern refs, constraints, slice hints (how)
- `plan.yml` — execution plan: slices (id, description, status, depends_on) containing tasks (id, task, status). Schema: `$AI_SESSION_HOME/spec/session/schemas/plan.schema.yml`.
- `questions.yml` — unresolved items
- `log.md` — append-only session history
- `review.yml`, `pr.md` — review and delivery artifacts

## Commands

- `/session:address-feedback` — Fetches and helps address feedback from the active PR.
- `/session:checkpoint` — Saves a checkpoint by updating state files.
- `/session:define` — Conversational session to define a new user story from scratch.
- `/session:end` — Ends the session, saving progress and project-wide knowledge to AGENTS.md.
- `/session:get-familiar` — Sub-agent summarizes the current branch's code changes.
- `/session:implement` — Executes all plan.yml tasks autonomously; verifies after each task; stops and logs on any unrecoverable failure. Headless only.
- `/session:log-research` — Logs a detailed summary of research findings to log.md.
- `/session:migration` — Migrates a legacy single-file feature document to the directory structure.
- `/session:new` — Creates a feature directory from a Shortcut story ID or Notion URL.
- `/session:plan` — Analyzes codebase and requirements to create a TDD-ready implementation plan.
- `/session:pr` — Generates a PR description, creates/updates the PR on GitHub.
- `/session:review` — Critical code review of the current branch using a sub-agent.
- `/session:review-devops` — DevOps-focused review of the current branch using a sub-agent.
- `/session:review-docs` — Documentation review of the current branch using a sub-agent.
- `/session:start` — Loads context from a feature directory to start or resume a session.
- `/session:summary` — Generates a human-readable Markdown summary of the feature's state.
- `/session:verify-release` — Verifies a cherry-picked release branch against original commits.

Primary entry points: `/session:define` (new story from scratch), `/session:new`
(from an existing ticket), `/session:start` (resume existing feature).

# Running Commands

Commands are invoked directly in the AI assistant's chat interface — not in the
terminal. Both tools use the same `/session:` prefix.

```
/session:start sc-12345
/session:plan
```

# Development Conventions

- **Command files:** `claude/session/*.md` is the single source of truth. Two generated outputs:
  - `gemini/session/*.toml` — generated via `scripts/gen_gemini.sh` (Gemini CLI commands). Do not edit directly.
  - `headless/session/*.md` — generated via `scripts/gen_headless.sh` (LLM-agnostic headless pipeline variants, Gemini tool names, no interactivity, no sub-agents). Do not edit directly except `plan.md` which is hand-written.
  Both scripts are incremental (checksum-based); use `--force` to regenerate all. Commit `.checksums` alongside generated files.
- **Scripts:** Shared helper scripts live in `scripts/` and are referenced via
  `$AI_SESSION_HOME/scripts/` in all commands. Added to `$PATH` via `setup.sh`.
  Key scripts:
  - `orchestrate.sh <story-id> --plan|--implement|--review|--pr` — invokes a single headless
    pipeline step with precondition checks and `status.yaml` writes. Run from inside the target repo.
  - `orchestrate.sh --status <story-id>` — prints raw `status.yaml` for a feature.
  - `test_orchestrate.sh` — smoke test for `orchestrate.sh` (arg parsing, preconditions, --status).
- **Session Context Pattern:** To reduce token consumption, session commands use an
  explicit context-passing pattern:
  - **Producers** (`/session:start`, `/session:define`, `/session:new`): Output a
    formatted block titled `### ✨ Session Context Loaded for...` containing the
    content of `description.md` and `AGENTS.md`.
  - **Consumers** (e.g., `/session:plan`, `/session:review`): Read the "Session
    Context" block from chat history instead of re-reading files from disk.
  - **Updaters** (e.g., `/session:pr`): If a command modifies a context file, it
    outputs a new updated "Session Context" block.
- **State Management:**
  - Unstructured: `description.md` and `log.md` (plain Markdown).
  - Structured: `plan.yml`, `questions.yml`, `review.yml` (YAML, modified via `ai-session` CLI).
  - All YAML modifications use `ai-session update-task` or `ai-session update-slice` for deterministic,
    atomic updates.
- **Command Patterns:**
  - **LLM Orchestrator:** For interactive tasks, the agent orchestrates helper scripts
    directly (e.g., `/session:define`).
  - **Sub-agent Pattern:** For focused one-off tasks, work is delegated to an isolated
    sub-agent to keep the main session clean (e.g., `/session:review`, `/session:pr`).

See `spec/session/README.md` for full architectural rationale and examples.

# Testing Conventions

Commands are tested by delegating execution to a sub-agent with the
`spec/session/example-feature-document/` directory as its working context. This
keeps the main session clean and simulates a realistic session environment.

Typical approach:
1. Ask the main agent to open a sub-agent for testing.
2. Instruct the sub-agent that it is running in a test environment using the
   example feature directory (`spec/session/example-feature-document/`) as the
   active feature.
3. Have the sub-agent execute the command under test and report the outcome.
4. Evaluate whether the output matches the expected behaviour described in
   `spec/session/command_details.md`.

## Verification
Run: make build && make test && make lint

