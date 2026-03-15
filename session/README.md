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
- `**/session:get_familiar`**: Gets familiar with the current code changes by having a subagent generate a summary.
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

---

## Command Details

This section provides a detailed breakdown of individual session commands, their dependencies, and their interactions with the file system and external tools.

### `/session:checkpoint`

-   **Description:** Saves a snapshot of the work-in-progress. It updates the status of tasks and questions based on the conversation history and logs a summary of the progress.
-   **Orchestration Pattern:** [LLM Orchestrator with Helper Scripts](#pattern-1-llm-orchestrator-with-helper-scripts)
-   **Dependencies:**
    -   **Skills:**
        -   `yq-skill`: Activated to get expert knowledge for constructing `yq` commands.
    -   **Scripts:**
        -   `scripts/append_to_log.sh`: Used to append a timestamped summary to the feature's log file.
    -   **Tools:**
        -   `run_shell_command`: Used to execute `yq` and the `append_to_log.sh` script.
-   **Interactions:**
    -   **Input (Reads):**
        -   Identifies the active feature directory (e.g., `.vscode/sc-XXXXX`) from the conversation context.
        -   Reads `plan.yml` to identify the status of tasks.
        -   Reads `questions.yml` to identify the status of questions.
    -   **Output (Writes):**
        -   Modifies `plan.yml` in-place to update the `status` of completed tasks to 'done'.
        -   Modifies `questions.yml` in-place to update the `status` of answered questions to 'resolved' and fill in the `answer`.
        -   Appends a generated Markdown summary to `log.md`.

### `/session:address-feedback`

-   **Description:** Fetches and helps address feedback comments from a feature`s GitHub Pull Request.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `read_file`, `pull_request_read`, `run_shell_command`
    -   **External Services:** GitHub
-   **Interactions:**
    -   **Input (Reads):** `description.md`, `GEMINI.md`, GitHub PR review comments
    -   **Output (Writes):** `log.md` (in the feature directory), project source files


### `/session:define`

-   **Description:** Starts a conversational session to define a new user story and create its feature directory.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/create_feature_dir.sh`
    -   **Tools:** `glob`, `grep_search`, `run_shell_command`, `write_file`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   `session/define.toml` (command definition)
        -   User command-line arguments (`{{args}}`)
        -   Project source code files (via `glob` and `grep_search`)
    -   **Output (Writes):**
        -   Creates a new feature directory (e.g., `.vscode/create-user-profile-page/`)
        -   `.vscode/<feature-name>/description.md`
        -   `.vscode/<feature-name>/plan.yml`
        -   `.vscode/<feature-name>/questions.yml`
        -   `.vscode/<feature-name>/review.yml`
        -   `.vscode/<feature-name>/log.md`
        -   `.vscode/<feature-name>/pr.md`


### `/session:end`

-   **Description:** Ends the work session, saving progress to a feature-specific directory and persisting project-wide knowledge.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** `yq YAML Processing`
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `read_file`, `run_shell_command`, `replace`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):** Session conversation history, `plan.yml`, `questions.yml`, `GEMINI.md`.
    -   **Output (Writes):** `plan.yml`, `questions.yml`, `log.md`, `GEMINI.md`.


### `/session:get_familiar`

-   **Description:** Uses a sub-agent to analyze and summarize the current Git branch`s code changes.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `generalist`
    -   **External Services:** Git remote (`origin`)
-   **Interactions:**
    -   **Input (Reads):** Local Git repository state (diff against the remote default branch, including committed, staged, and unstaged files).
    -   **Output (Writes):** Writes a summary of code changes to standard output.


### `/session:log-research`

-   **Description:** Logs a detailed, comprehensive summary of research findings to a feature-specific `log.md` file.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `/scripts/append_to_log.sh`
    -   **Tools:** `run_shell_command`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):** Reads the LLM's conversation history and the target log file to check its state (e.g., if it`s empty).
    -   **Output (Writes):** Appends a timestamped report to a feature-specific log file (e.g., `./.vscode/sc-XXXXX/log.md`).


### `/session:migration`

-   **Description:** Migrates a legacy, single-file feature markdown document into the modern, multi-file directory structure.
-   **Orchestration Pattern:** Shell Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** 
        - `/scripts/migrate_feature_file.sh`
    -   **Tools:** `chmod`, `mkdir`, `touch`, `awk`, `mv`, `sed`, `basename` (all used within the script).
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   A single markdown file path provided as an argument (e.g., `/path/to/feature.md`).
    -   **Output (Writes):**
        -   Creates a new directory named after the input file (e.g., `/path/to/feature/`).
        -   Writes parsed content into the following new files within that directory:
            -   `description.md`
            -   `plan.yml`
            -   `questions.yml`
            -   `review.yml`
            -   `log.md`
            -   `pr.md`
        -   Archives the original input file by renaming it with a `.migrated` extension.


### `/session:new`

-   **Description:** Creates a feature directory based on a Shortcut story ID, fetching related resources to populate a description file.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/create_feature_dir.sh`
    -   **Tools:** `run_shell_command`, `stories_get_by_id`, `notion_fetch`, `write_file`
    -   **External Services:** Shortcut, Notion
-   **Interactions:**
    -   **Input (Reads):**
        -   Shortcut Story from the Shortcut API.
        -   Notion Page from the Notion API.
    -   **Output (Writes):**
        -   `.vscode/<story-id>/description.md`
        -   `.vscode/<story-id>/plan.yml`
        -   `.vscode/<story-id>/questions.yml`
        -   `.vscode/<story-id>/review.yml`
        -   `.vscode/<story-id>/log.md`
        -   `.vscode/<story-id>/pr.md`


### `/session:plan`

-   **Description:** Analyzes codebase and feature requirements to create a detailed, TDD-ready implementation plan.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** None
    -   **Tools:** `read_file`, `glob`, `grep_search`, `write_file`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   `{feature_directory}/description.md`
        -   `GEMINI.md`
        -   Codebase files via `glob` and `grep_search`.
    -   **Output (Writes):**
        -   `plan.yml`
        -   `questions.yml`


### `/session:pr_from_branch`

-   **Description:** Generates a pull request description by analyzing git changes and optionally fetching context from a Shortcut story linked in the branch name.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `run_shell_command`, `read_file`, `write_file`, `mcp_shortcut_stories_get_by_id`, `mcp_github_search_pull_requests`, `mcp_github_update_pull_request`, `mcp_github_create_pull_request`
    -   **External Services:** GitHub, Shortcut
-   **Interactions:**
    -   **Input (Reads):** Local git repository (current branch, diff from default branch), `.vscode/pull_request_template.md`, Shortcut API (story details), GitHub API (searches for open PRs).
    -   **Output (Writes):** GitHub API (creates or updates a pull request), `pull_request_descr.md` (as a fallback).


### `/session:pr`

-   **Description:** Generates a pull request description, creates or updates the PR on GitHub, and saves the resulting PR link to the feature directory.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `run_shell_command`, `search_pull_requests`, `read_file`, `create_pull_request`, `update_pull_request`, `write_file`, `ask_user`
    -   **External Services:** GitHub
-   **Interactions:**
    -   **Input (Reads):** Git context from `get_git_context.sh` (including branch and diff), `.vscode/pull_request_template.md`, feature directory files (`description.md`, `plan.yml`, `log.md`), existing pull request data from GitHub.
    -   **Output (Writes):** A pull request on GitHub, `description.md` within the active feature directory, or `pull_request_descr.md` in the project root as a fallback.


### `/session:review_from_branch`

-   **Description:** Performs a context-aware code review of the current branch against its associated Shortcut story.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `/scripts/get_git_context.sh`
    -   **Tools:** `stories_get_by_id`, `read_file`, `grep_search`, `stories_create_comment`
    -   **External Services:** Shortcut
-   **Interactions:**
    -   **Input (Reads):** Local git repository state (via script), `GEMINI.md`, project source files, Shortcut API.
    -   **Output (Writes):** Shortcut API (story comments).


### `/session:review`

-   **Description:** Performs a critical, context-aware code review of the current branch.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:**
        -   `scripts/get_git_context.sh`
    -   **Tools:**
        -   `run_shell_command`
        -   `read_file`
        -   `write_file`
    -   **External Services:**
        -   Git Remote (e.g., GitHub, GitLab)
-   **Interactions:**
    -   **Input (Reads):**
        -   Git repository state (branch, diff) via `scripts/get_git_context.sh`.
        -   `description.md` from the active feature directory.
        -   `GEMINI.md` from the project root.
    -   **Output (Writes):**
        -   `review.yml` to the active feature directory.


### `/session:start`

-   **Description:** Starts a work session by loading context files from a specified feature directory and passing them to a new, focused sub-session.
-   **Orchestration Pattern:** Shell Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `/scripts/load_context_files.sh`
    -   **Tools:** `gemini query`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   Reads multiple files from the feature directory (`.vscode/{{args}}`): `description.md`, `plan.yml`, `questions.yml`, `review.yml`, `log.md`, `pr.md`.
        -   Reads the global project documentation file: `GEMINI.md`.
    -   **Output (Writes):**
        -   Pipes the consolidated content of all read files into a `gemini query` sub-session for review.


### `/session:summary`

-   **Description:** Generates a human-readable Markdown summary of a feature`s state from its constituent context files.
-   **Orchestration Pattern:** Shell Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/load_context_files.sh`
    -   **Tools:** `write_file`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):**
        -   `scripts/load_context_files.sh`
        -   `.vscode/{{args}}/description.md`
        -   `.vscode/{{args}}/plan.yml`
        -   `.vscode/{{args}}/questions.yml`
        -   `.vscode/{{args}}/review.yml`
        -   `.vscode/{{args}}/log.md`
        -   `.vscode/{{args}}/pr.md`
        -   `GEMINI.md`
    -   **Output (Writes):**
        -   `.vscode/{{args}}/_SUMMARY.md`


### `/session:verify-release`

-   **Description:** Verifies a cherry-picked release branch against its original commits and provides an AI analysis of any discrepancies.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/verify-release.sh`
    -   **Tools:** `run_shell_command`
    -   **External Services:** None
-   **Interactions:**
    -   **Input (Reads):** Local git repository (branches, commits, commit messages, and patch data).
    -   **Output (Writes):** Temporary patch files for comparison. Standard output for verification results.

