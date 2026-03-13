# Project Overview

This project is a suite of custom Gemini CLI commands designed to implement a structured, session-based workflow for software development. The commands orchestrate a lifecycle for feature development, from creation and planning to review and pull request generation.

The core of the workflow revolves around a "feature directory" (a directory stored in `.vscode/`) which acts as a single source of truth for a given task, and this `GEMINI.md` file, which provides global project context.

The defined commands are:
- `**/session:address-feedback**: Fetches and helps address feedback comments from the active feature's GitHub Pull Request.
- `**/session:checkpoint**: Saves a checkpoint of the work done by updating the state files using the yq tool.
- `**/session:define**: Starts a conversational session to define a new user story and create its feature directory.
- `**/session:end**: Ends the work session, saving progress to the feature directory and project-wide knowledge to GEMINI.md.
- `**/session:log-research**: Logs a detailed, comprehensive summary of research findings to log.md.
- `**/session:migration**: Migrates an old, single-file feature document to the new directory structure with structured YAML files.
- `**/session:new**: Creates a new feature directory from a Shortcut story ID.
- `**/session:plan**: Analyzes codebase and feature requirements to create a detailed, TDD-ready implementation plan.
- `**/session:pr**: Generates a pull request description, creates/updates the PR on GitHub, and saves the link to the feature directory.
- `**/session:pr_from_branch**: Generates a PR description. If branch name has a Shortcut story, it uses it for context and links the PR back to the story.
- `**/session:prepare-release**: Prepares a release by creating a branch, cherry-picking commits, and intelligently resolving any conflicts.
- `**/session:review**: Performs a critical, context-aware code review of the current branch.
- `**/session:review_from_branch**: Performs a critical, context-aware code review of the current branch, using the Shortcut story from the branch name as context.
- `**/session:start**: Starts a work session by loading context from a feature directory and the project's GEMINI file.
- `**/session:summary**: Generates a human-readable Markdown summary of the entire feature's state.
- `**/session:verify-release**: Verifies a cherry-picked release on the current branch, providing an AI-powered analysis of any changes found.

# Building and Running

This project requires the `yq` command-line tool (v4+) to be installed and available in the system's PATH.

The commands are executed directly within the Gemini CLI. For example:
```bash
/session:start sc-12345
```
Followed by:
```bash
/session:plan
```

# Development Conventions

*   **Command Definition**: Each command is defined in a `.toml` file within the `session/` directory.
*   **Prompt-Based Logic**: The core logic for each command is contained within its `prompt` field.
*   **State Management**: The workflow state is managed through the contents of the feature directory files.
    *   **Unstructured Data:** `description.md` and `log.md` are standard markdown files.
    *   **Structured Data:** `plan.yml`, `questions.yml`, and `review.yml` are structured YAML files.
    *   **Modification Pattern:** All modifications to these YAML files are performed by activating the `yq-skill` and using `run_shell_command` to execute `yq` commands. This provides atomic, deterministic, and robust state updates, which is the core principle of this project's architecture.
*   **Skills**: The workflow relies on locally installed skills (`tdd-skill`, `yq-skill`) for complex, reusable logic.
*   **Command Prompt Design**: Commands can be implemented in two main ways:
    *   **Shell Script Prompt**: For simple, deterministic, non-interactive tasks, the `prompt` can be a `#!/bin/bash` script.
    *   **LLM Orchestrator Prompt**: For complex workflows that require loops, conditional logic, error handling, or AI-powered analysis, the `prompt` should be a natural language set of instructions for the agent, which then uses tools to execute the steps.
*   **Hybrid Pattern (Preferred)**: The most robust pattern is to use an **LLM Orchestrator Prompt** as the main entry point for a command. This orchestrator should then call small, single-purpose **helper scripts** (located in `scripts/`) for any procedural or deterministic steps where precision is critical (e.g., generating a timestamped file name). This balances the reliability of scripts for predictable tasks with the flexibility of the LLM for complex, interactive ones. Commands like `/session:new`, `/session:checkpoint`, and `/session:prepare-release` are good examples of this pattern. See the `session/README.md` for the full architectural rationale.

# Skill Development

(No changes in this section)
...
