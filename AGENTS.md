# Project Overview

This project is a suite of AI assistant commands that implement a structured,
session-based workflow for software development, compatible with both Gemini CLI and
Claude Code. The commands orchestrate a lifecycle for feature development, from
creation and planning to review and pull request generation.

The core of the workflow revolves around a "feature directory" (stored in `.vscode/`
by default) which acts as a single source of truth for a given task, and an `AGENTS.md`
file in each project which provides global context. See
`spec/session/example-feature-document/` for a complete example of a feature directory.

## Repo Structure

```
gemini/session/   ← Gemini CLI implementation (*.toml)
claude/session/   ← Claude Code implementation (*.md)
spec/session/     ← LLM-agnostic documentation and examples
scripts/          ← Shared shell scripts used by both tools
```

## Commands

- `/session:address-feedback` — Fetches and helps address feedback from the active PR.
- `/session:checkpoint` — Saves a checkpoint by updating state files.
- `/session:define` — Conversational session to define a new user story from scratch.
- `/session:end` — Ends the session, saving progress and project-wide knowledge to AGENTS.md.
- `/session:get-familiar` — Sub-agent summarizes the current branch's code changes.
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

- **Command files:** Gemini commands are `.toml` files in `gemini/session/`. Claude
  commands are `.md` files in `claude/session/`. Both implement the same concepts.
- **Scripts:** Shared helper scripts live in `scripts/` and are referenced via
  `$AI_SESSION_HOME/scripts/` in all commands.
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
  - Structured: `plan.yml`, `questions.yml`, `review.yml` (YAML, modified via `yq`).
  - All YAML modifications use `yq` commands executed via shell for deterministic,
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
