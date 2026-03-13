# Session Workflow Commands

This document describes a suite of custom Gemini CLI commands designed to implement a structured, session-based workflow for software development. The workflow is centered around a "feature directory" that acts as a single source of truth for a task.

## Project Philosophy

The primary goal of this command suite is to create a robust, semi-automated development loop that intelligently manages context. The core idea is to differentiate between two types of knowledge:

1.  **Feature-Specific Knowledge:** The details, requirements, implementation plan, and progress related to a single user story. This context is ephemeral and only relevant for the duration of the task.
2.  **Project-Wide Knowledge:** Enduring architectural patterns, conventions, and learnings that should persist and be shared across all future development in the repository.

This suite of commands orchestrates the flow of information between the user, the codebase, external tools, and dedicated documents that store these two types of knowledge, creating a persistent "memory" for both the specific task and the overall project.

## Core Concepts

-   **Feature Directory:** A directory located in `.vscode/` (e.g., `.vscode/sc-12345/`). It contains a mix of Markdown files (like `description.md`, `log.md`) and structured YAML files (`plan.yml`, `questions.yml`, `review.yml`) that hold the state for a specific feature. This serves as the "session memory."
-   **Project Document (`GEMINI.md`):** A global file that stores project-wide context, architectural guidelines, and conventions. This serves as the "project memory."

## Dependencies

This workflow has the following dependencies:

-   **`yq` command-line tool (v4+):** Required for atomic and reliable modification of `.yml` state files. It must be installed and available in the system's PATH.
-   **`yq-skill` & `tdd-skill`:** Locally installed Gemini skills that provide expert knowledge.
-   **External Services:** Integrations with Shortcut, Notion, Git, and GitHub are used for various commands.

## Design Notes & Conventions

### A Note on the `.vscode` Directory

The choice to store feature artifacts in `.vscode/` is a practical one based on personal habit. Because the `.vscode` folder is ignored in many of our projects, I have been using it to keep personal, project-related files that I don't want to commit.

### Architectural Rationale

This project has evolved through several stages, with each step aimed at increasing reliability and using the best tool for the job. The core principle is to use deterministic, specialized tools (`yq`, shell scripts) for procedural tasks, and to reserve the LLM for creative, analytical, and orchestrating tasks. This has led to two primary architectural patterns for creating commands.

#### Pattern 1: LLM Orchestrator with Helper Scripts
This pattern is ideal for complex, interactive commands that may require conditional logic, loops, or user interaction.

The pattern is as follows:
1.  The command's `prompt` is defined as a **high-level LLM prompt**, not a shell script. This prompt makes the agent an "orchestrator" for the entire command workflow.
2.  The orchestrator agent uses the `run_shell_command` tool to execute small, deterministic, single-purpose **helper scripts** for predictable steps where precision is critical (e.g., generating a filename with a specific timestamp format, running git commands).
3.  The orchestrator agent then handles the complex, stateful, or interactive parts of the workflow itself, using its reasoning capabilities to manage the process.

This architecture balances the reliability of scripts for deterministic tasks with the analytical flexibility of the LLM for complex ones. Commands like `/session:define` and `/session:review` (before its refactoring) are good examples.

#### Pattern 2: Shell Orchestrator with Focused LLM Sub-sessions
This pattern (also called "Gemini Inception") is the most advanced and efficient architecture in the suite. It is ideal for commands that need to perform a specific, one-off AI task (like summarization or generation) on a large amount of data without polluting the main session context.

The pattern is as follows:
1.  The command's `prompt` is defined as a `#!/bin/bash` **shell script**, which acts as the main orchestrator.
2.  This script first gathers and prepares a minimal, focused context for the task, often by calling other helper scripts (e.g., `scripts/load_context_files.sh`).
3.  The orchestrator script then delegates the AI-heavy task to a temporary, isolated **sub-session** by piping the prepared context directly into a `gemini query "..."` command.
4.  The orchestrator script captures the output of the sub-session and performs any final actions. The temporary sub-session and its large context are destroyed upon completion.

This provides maximum efficiency, context isolation, and token economy. The `/session:start` and `/session:summary` commands are the canonical examples of this pattern.

---

## Commands

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
- `**/session:review**: Performs a critical, context-aware code review of the current branch.
- `**/session:review_from_branch**: Performs a critical, context-aware code review of the current branch, using the Shortcut story from the branch name as context.
- `**/session:start**: Starts a work session by loading context from a feature directory and the project's GEMINI file.
- `**/session:summary**: Generates a human-readable Markdown summary of the entire feature's state.
- `**/session:verify-release**: Verifies a cherry-picked release on the current branch, providing an AI-powered analysis of any changes found.
