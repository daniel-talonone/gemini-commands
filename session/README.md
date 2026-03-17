# Session Workflow Commands

This document describes a suite of custom Gemini CLI commands designed to implement a structured, session-based workflow for software development. The workflow is centered around a "feature directory" that acts as a single source of truth for a task.

## Project Philosophy

The primary goal of this command suite is to create a robust, semi-automated development loop that intelligently manages context. The core idea is to differentiate between two types of knowledge:

1.  **Feature-Specific Knowledge:** The details, requirements, implementation plan, and progress related to a single user story. This context is ephemeral and only relevant for the duration of the task.
2.  **Project-Wide Knowledge:** Enduring architectural patterns, conventions, and learnings that should persist and be shared across all future development in the repository.

This suite of commands orchestrates the flow of information between the user, the codebase, external tools, and dedicated documents that store these two types of knowledge, creating a persistent "memory" for both the specific task and the overall project.

## Core Concepts

-   **Feature Directory:** A directory located in `.vscode/` (e.g., `.vscode/sc-12345/`). It contains a mix of Markdown files (like `description.md`, `log.md`) and structured YAML files (`plan.yml`, `questions.yml`, `review.yml`) that hold the state for a specific feature. This serves as the "session memory." See the `example-feature-document/` directory for a complete example.
-   **Project Document (`GEMINI.md`):** A global file that stores project-wide context, architectural guidelines, and conventions. This serves as the "project memory."

## Getting Started: Session Entry Points

To begin a work session, there are three primary commands, each serving a distinct purpose:

*   **`/session:define`**: Use this to define a **new user story from scratch**. This command guides you through a conversational process to capture requirements and creates a new feature directory.
*   **`/session:new`**: Use this to **create a feature document from an existing user story ID or Notion page URL**. This command fetches information from external services (like Shortcut or Notion) to pre-populate your feature directory.
*   **`/session:start`**: Use this to **resume work on an existing feature**. This command loads all context from a previously created feature directory into your current session.

Once a session is started (either with `define`, `new`, or `start`), you can proceed with planning, implementation, and other workflow commands.

## Dependencies

This workflow has the following dependencies:

-   **`yq` command-line tool (v4+):** Required for atomic and reliable modification of `.yml` state files. It must be installed and available in the system's PATH.
-   **`yq-skill` & `tdd-skill`:** Locally installed Gemini skills that provide expert knowledge.
-   **External Services:** Integrations with Shortcut, Notion, Git, and GitHub are used for various commands. (Note: Specific API versions are used for stability; refer to individual command details for specifics if applicable.)

## Design Notes & Conventions

### A Note on the `.vscode` Directory

The choice to store feature artifacts in `.vscode/` is a practical one based on personal habit. Because the `.vscode` folder is ignored in many of our projects, I have been using it to keep personal, project-related files that I don't want to commit.

### Architectural Rationale

This project has evolved through several stages, with each step aimed at increasing reliability and using the best tool for the job. The core principle is to use deterministic, specialized tools (`yq`, shell scripts) for procedural tasks, and to reserve the LLM for creative, analytical, and orchestrating tasks. This has led to the following architectural patterns:

#### LLM Orchestrator with Helper Scripts
This pattern is ideal for complex, interactive commands that may require conditional logic, loops, or user interaction.

The pattern is as follows:
1.  The command's `prompt` is defined as a **high-level LLM prompt**, not a shell script. This prompt makes the agent an "orchestrator" for the entire command workflow.
2.  The orchestrator agent uses the `run_shell_command` tool to execute small, deterministic, single-purpose **helper scripts** for predictable steps where precision is critical (e.g., generating a filename with a specific timestamp format, running git commands).
3.  The orchestrator agent then handles the complex, stateful, or interactive parts of the workflow itself, using its reasoning capabilities to manage the process.

This architecture balances the reliability of scripts for deterministic tasks with the analytical flexibility of the LLM for complex ones. Commands like `/session:define` and `/session:review` are good examples.

#### Subagent Pattern for Focused Tasks
This is the modern and preferred pattern for delegating a complex, one-off task to an isolated environment.

The pattern is as follows:
1.  The command's `prompt` instructs the main agent to use the `generalist` sub-agent tool.
2.  The prompt includes a detailed set of instructions that will be passed to the `generalist`.
3.  The main agent calls the `generalist`, which executes the task in a completely isolated session with its own context and tools.
4.  The final result is returned to the main agent.

This provides maximum efficiency and context isolation. Commands like `/session:get-familiar`, `/session:checkpoint`, and `/session:end` are good examples of this pattern.

---

## Commands

- `**/session:address-feedback**: Fetches and helps address feedback comments from the active feature's GitHub Pull Request.
- `**/session:checkpoint**: Saves a checkpoint of the work done by updating the state files using the yq tool.
- `**/session:define**: Starts a conversational session to define a new user story and create its feature directory.
- `**/session:end**: Ends the work session, saving progress to the feature directory and project-wide knowledge to GEMINI.md.
- `**/session:get-familiar`**: Gets familiar with the current code changes by having a subagent generate a summary.
- `**/session:log-research**: Logs a detailed, comprehensive summary of research findings to log.md.
- `**/session:migration**: Migrates an old, single-file feature document to the new directory structure with structured YAML files.
- `**/session:new**: Creates a new feature directory from a Shortcut story ID or Notion page URL.
- `**/session:plan**: Analyzes codebase and feature requirements to create a detailed, TDD-ready implementation plan.
- `**/session:pr**: Generates a pull request description, creates/updates the PR on GitHub, and saves the link to the feature directory.
- `**/session:review**: Performs a critical, context-aware code review of the current branch using a focused sub-agent.
- `**/session:start**: Starts a work session by loading context from a feature directory and the project's GEMINI file.
- `**/session:summary**: Generates a human-readable Markdown summary of the entire feature's state.
- `**/session:verify-release**: Verifies a cherry-picked release on the current branch, providing an AI-powered analysis of any changes found.

---

## Command Details

This section provides a detailed breakdown of individual session commands, their dependencies, and their interactions with the file system and external tools.

### `/session:address-feedback`

-   **Description:** Fetches and helps address unresolved review comments from a feature's GitHub Pull Request.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `read_file`, `pull_request_read`, `run_shell_command`
    -   **External Services:** GitHub
-   **Interactions:**
    -   **Input (Reads):**
        -   `.vscode/<feature-dir>/description.md` (to get PR URL)
        -   `GEMINI.md`
        -   GitHub API (to get review comments).
    -   **Output (Writes):**
        -   Appends a summary to `.vscode/<feature-dir>/log.md`.
        -   Modifies project source files to address feedback.

### `/session:checkpoint`

-   **Description:** Saves a snapshot of the work-in-progress by updating the status of tasks and questions and logging a summary of the progress.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Skills:** `yq YAML Processing`
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `run_shell_command`, `generalist`
-   **Interactions:**
    -   **Input (Reads):**
        -   The active feature directory path from the conversation.
        -   `.vscode/<feature-dir>/plan.yml`
        -   `.vscode/<feature-dir>/questions.yml`
    -   **Output (Writes):**
        -   Modifies `.vscode/<feature-dir>/plan.yml` in-place.
        -   Modifies `.vscode/<feature-dir>/questions.yml` in-place.
        -   Appends a summary to `.vscode/<feature-dir>/log.md`.

### `/session:define`

-   **Description:** Starts a conversational session to define a new user story and creates the corresponding feature directory and artifacts.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/create_feature_dir.sh`
    -   **Tools:** `glob`, `grep_search`, `run_shell_command`, `write_file`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   User's command-line arguments (`{{args}}`).
        -   Project source code via `glob` and `grep_search`.
    -   **Output (Writes):**
        -   Creates a new feature directory (e.g., `.vscode/create-user-profile-page/`).
        -   `.vscode/<feature-dir>/description.md`
        -   `.vscode/<feature-dir>/plan.yml`
        -   `.vscode/<feature-dir>/questions.yml`
        -   `.vscode/<feature-dir>/review.yml`
        -   `.vscode/<feature-dir>/log.md`
        -   `.vscode/<feature-dir>/pr.md`

### `/session:end`

-   **Description:** Ends the work session, saving a final summary to the feature directory and persisting any project-wide knowledge to `GEMINI.md`.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Skills:** `yq YAML Processing`
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `read_file`, `run_shell_command`, `replace`, `generalist`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   Session conversation history.
        -   `.vscode/<feature-dir>/plan.yml`
        -   `.vscode/<feature-dir>/questions.yml`
        -   `GEMINI.md`
    -   **Output (Writes):**
        -   Modifies `.vscode/<feature-dir>/plan.yml` in-place.
        -   Modifies `.vscode/<feature-dir>/questions.yml` in-place.
        -   Appends a final summary to `.vscode/<feature-dir>/log.md`.
        -   Modifies `GEMINI.md` in-place.

### `/session:get-familiar`

-   **Description:** Uses a sub-agent to analyze and summarize the current Git branch's code changes.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `generalist`
    -   **External Services:** Git
-   **Interactions:**
    -   **Input (Reads):**
        -   Local Git repository state (diff against the remote default branch).
    -   **Output (Writes):**
        -   Writes a summary of code changes to standard output.

### `/session:log-research`

-   **Description:** Logs a detailed, comprehensive summary of research findings to the feature's `log.md` file.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `run_shell_command`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   Session conversation history.
    -   **Output (Writes):**
        -   Appends a timestamped report to `.vscode/<feature-dir>/log.md`.

### `/session:migration`

-   **Description:** Migrates a legacy, single-file feature markdown document into the modern, multi-file directory structure.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/migrate_feature_file.sh`
    -   **Tools:** `run_shell_command`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   A single markdown file path provided as an argument.
    -   **Output (Writes):**
        -   Creates a new directory named after the input file.
        -   Populates the new directory with `description.md`, `plan.yml`, `questions.yml`, etc.
        -   Archives the original file by renaming it with a `.migrated` extension.

### `/session:new`

-   **Description:** Creates a feature directory from a Shortcut story ID or Notion page URL, fetching related resources to populate the `description.md` file.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/create_feature_dir.sh`
    -   **Tools:** `run_shell_command`, `stories_get_by_id`, `notion_fetch`, `write_file`
    -   **External Services:** Shortcut, Notion
-   **Interactions:**
    -   **Input (Reads):**
        -   Shortcut API (to get story details if a story ID is provided).
        -   Notion API (to get page content if a Notion URL is provided).
    -   **Output (Writes):**
        -   Creates a new directory and populates it with `description.md`, `plan.yml`, etc.

### `/session:plan`

-   **Description:** Analyzes codebase and feature requirements to create a detailed, TDD-ready implementation plan.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** None
    -   **Tools:** `read_file`, `glob`, `grep_search`, `write_file`
    -   **External Services:** None
-   **Interactions:** (Orchestrates by directly using Gemini CLI tools)
    -   **Input (Reads):**
        -   `.vscode/<feature-dir>/description.md`
        -   `GEMINI.md`
        -   Codebase files via `glob` and `grep_search`.
        -   User input during interactive planning.
    -   **Output (Writes):**
        -   Generates and writes initial `.vscode/<feature-dir>/plan.yml`
        -   Generates and writes initial `.vscode/<feature-dir>/questions.yml`

### `/session:pr`

-   **Description:** Generates a pull request description, creates or updates the PR on GitHub, and saves the resulting PR link to the feature directory.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `run_shell_command`, `search_pull_requests`, `read_file`, `create_pull_request`, `update_pull_request`, `write_file`, `ask_user`
    -   **External Services:** GitHub
-   **Interactions:**
    -   **Input (Reads):**
        -   Git repository state (via script).
        -   `.vscode/pull_request_template.md`
        -   Feature directory files (`description.md`, `plan.yml`, `log.md`).
        -   GitHub API (to search for existing PRs).
    -   **Output (Writes):**
        -   Creates or updates a pull request on GitHub.
        -   Writes the PR link to `.vscode/<feature-dir>/description.md`.
        -   `pull_request_descr.md` (as a fallback).

### `/session:review`

-   **Description:** Performs a critical, context-aware code review of the current branch using a focused sub-agent.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `run_shell_command`, `read_file`, `generalist`
    -   **External Services:** Git
-   **Interactions:**
    -   **Input (Reads):**
        -   Git repository state (via script).
        -   `.vscode/<feature-dir>/description.md`
        -   `GEMINI.md`
        -   Reads back the `.vscode/<feature-dir>/review.yml` for verification.
    -   **Output (Writes):**
        -   Delegates writing `.vscode/<feature-dir>/review.yml` to a sub-agent.

### `/session:start`

-   **Description:** Starts a work session by loading all context from a feature directory and the project's `GEMINI.md` file.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/load_context_files.sh`
    -   **Tools:** `run_shell_command`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   The output of the `load_context_files.sh` script, which contains the content of all files in the feature directory and `GEMINI.md`.
    -   **Output (Writes):**
        -   The context is loaded into the main session. The command confirms completion to the user.

### `/session:summary`

-   **Description:** Generates a single, human-readable Markdown summary of the feature's entire state.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/load_context_files.sh`
    -   **Tools:** `run_shell_command`, `write_file`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   The output of the `load_context_files.sh` script, which contains the content of all files in the feature directory and `GEMINI.md`.
    -   **Output (Writes):**
        -   `.vscode/<feature-dir>/_SUMMARY.md`

### `/session:verify-release`

-   **Description:** Verifies a cherry-picked release branch against its original commits and provides an AI-powered analysis of any discrepancies.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/verify-release.sh`
    -   **Tools:** `run_shell_command`
    -   **External Services:** Git
-   **Interactions:**
    -   **Input (Reads):**
        -   Local Git repository (branches, commits, commit messages, and patch data).
    -   **Output (Writes):**
        -   Writes temporary patch files for comparison.
        -   Writes a verification report to standard output.
