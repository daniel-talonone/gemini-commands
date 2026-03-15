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
- `**/session:review**: Performs a critical, context-aware code review of the current branch.
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
    *   **Modification Pattern:** All modifications to these YAML files are performed by activating the `yq-skill` and using `run_shell_command` to execute `yq` commands. This provides atomic, deterministic, and robust state updates.
*   **Skills**: The workflow relies on locally installed skills (`tdd-skill`, `yq-skill`) for complex, reusable logic.
*   **Command Patterns**: Commands are implemented using the **LLM Orchestrator** pattern. For complex, interactive tasks (e.g., `/session:define`), the `prompt` is a high-level set of instructions for the agent. The agent acts as an orchestrator, using tools like `run_shell_command` to call helper scripts for deterministic steps, while managing the overall workflow and user interaction. For delegating tasks to a sub-agent, the `generalist` tool is used (e.g., `/session:get_familiar`).
*   See the `session/README.md` for the full architectural rationale and detailed examples of these patterns.

# Skill Development
...
