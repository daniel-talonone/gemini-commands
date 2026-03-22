# A Session-Based Workflow for Development

> **Note on Active Development:** This project is under active development and is used as a playground for exploring and experimenting with AI-assisted development concepts. As such, commands and workflows may change or break unexpectedly.

> **Important Note on Feature Directories:** By default, feature directories are located in `.vscode/`. However, this can be overridden by a custom path defined in your project's `GEMINI.md`. For more details, refer to [A Note on the `.vscode` Directory](#a-note-on-the-vscode-directory).

## Overview

This document describes a suite of custom Gemini CLI commands for a structured, session-based workflow. The goal is to provide a semi-automated development loop that manages context by separating two types of knowledge:

1.  **Feature-Specific Knowledge:** The details, requirements, implementation plan, and progress related to a single user story. This context is ephemeral and only relevant for the duration of the task.
2.  **Project-Wide Knowledge:** Architectural patterns, conventions, and learnings that should persist and be shared across all future development in the repository.

This suite of commands orchestrates the flow of information between the user, the codebase, external tools, and dedicated documents that store these two types of knowledge. This creates persistent storage for both task-specific and project-wide information.


## Core Concepts

-  **Session:** A session represents the working context for a single feature or user story.

    In this workflow, a session is not only tied to the terminal or chat history. Instead, it is primarily defined by the feature directory, which stores the description, implementation plan, open questions, progress log, and review notes for the feature.

    The terminal session is still used during development, but the feature directory acts as a more stable source of context. This makes it easier to resume work across multiple days and allows the LLM to use structured information rather than incomplete conversational history.
-   **Feature Directory:** A directory located in `.vscode/` (e.g., `.vscode/sc-12345/`). It contains a mix of Markdown files (like `description.md`, `log.md`) and structured YAML files (`plan.yml`, `questions.yml`, `review.yml`) that hold the state for a specific feature. This serves as the "session memory." See the `example-feature-document/` directory for a complete example.

    Example:
    ```
    .vscode/sc-12345/
        description.md
        plan.yml
        questions.yml
        log.md
        review.yml
    ```
-   **Project Document (`GEMINI.md`):** A global file that stores project-wide context, architectural guidelines, and conventions. This serves as the "project memory" and helps maintain and apply project knowledge consistently across all sessions.

    Examples of knowledge to include in `GEMINI.md`:
    -   Architectural conventions and rules
    -   Testing standards and patterns
    -   Recurring development patterns
    -   Project-specific terminology and definitions.

## Evolution of the Approach

This project has evolved through several stages, with each step aiming to improve the process of AI-assisted development:

*   **Phase 1: Prompt-based** - Initial approach with no structured context, reliant on chat history.
*   **Phase 2: Single Feature Document** - Centralized all feature information into one document, which became unwieldy.
*   **Phase 3: Feature Directory** - Split the single document into multiple, domain-specific files within a feature directory.
*   **Phase 4: YAML + yq, Scripts, Subagents** - Introduced structured YAML files for state management, `yq` for updates, and helper scripts and subagents for deterministic task execution.

This evolution represents a shift from a conversation-driven to a state-driven workflow, where explicit, structured state takes precedence over ephemeral chat history.

## Getting Started: Session Entry Points

To begin a work session, there are three primary commands:

*   **`/session:define`**: Use this to define a **new user story from scratch**. This command guides you through a conversational process to capture requirements and creates a new feature directory.
*   **`/session:new`**: Use this to **create a feature document from an existing user story ID or Notion page URL**. This command fetches information from external services (like Shortcut or Notion) to pre-populate your feature directory.
*   **`/session:start`**: Use this to **resume work on an existing feature**. This command loads context from a previously created feature directory into your current session.

Once a session is started, you can proceed with other workflow commands.

## Workflow Lifecycle

While the commands can be used flexibly, they are designed to support a typical feature development lifecycle. The process is initiated by one of the entry-point commands and then proceeds through planning, implementation, and delivery.

Here is a typical workflow:

1.  **Start a Session** (using one of the [entry points](#getting-started-session-entry-points)).
2.  **Plan the Work**:
    *   `/session:plan`: The agent analyzes the requirements and codebase to produce an implementation plan (`plan.yml`).
3.  **Implement**:
    *   The developer writes the code to implement the tasks defined in the plan.
4.  **Track Progress** (Optional, as needed):
    *   `/session:checkpoint`: Save a snapshot of the work, update task statuses, and log progress.
    *   `/session:log-research`: Add research notes to the session log.
5.  **Review and Deliver**:
    *   `/session:review`: The agent performs a code review of the local changes.
    *   `/session:pr`: The agent generates a pull request description and creates the PR on GitHub.
    *   `/session:address-feedback`: After a PR is created, this command helps to fetch and address any unresolved review comments.
6.  **End the Session**:
    *   `/session:end`: The agent saves the final state and any project-wide learnings.

This lifecycle helps capture and utilize context, from initial requirements to final delivery.

**Note**: The lifecycle described above is an example path. The system is designed for flexibility. You can run `/session:end` at any point to store the current state and later resume your work with `/session:start`.

## Dependencies

-   **`yq` command-line tool (v4+):** Used for modifying `.yml` state files. It must be installed and available in the system's PATH.
-   **[`yq-skill`](https://mcpmarket.com/tools/skills/yaml-processing-transformation) & `tdd-skill`:** Locally installed Gemini skills.
-   **External Services:** Integrations with Shortcut, Notion, Git, and GitHub are used for various commands. For documentation on the available commands and their usage, refer to the Gemini CLI documentation for [Shortcut](https://www.shortcut.com/blog/why-we-built-the-shortcut-mcp-server), [Notion](https://developers.notion.com/guides/mcp/get-started-with-mcp), [Git](https://pypi.org/project/mcp-server-git/), and [GitHub](https://github.com/github/github-mcp-server/blob/main/docs/installation-guides/install-gemini-cli.md).
    To enable these integrations, configure your `.gemini/settings.json` with the following MCP servers:
    ```json
    "shortcut": {
        "command": "npx",
        "args": [
        "-y",
        " @shortcut/mcp@latest"
        ],
        "env": {
        "SHORTCUT_API_TOKEN": "{SHORTCUT_TOKEN}"
        }
    },
    "notion": {
        "command": "npx",
        "args": [
        "-y",
        "mcp-remote",
        "https://mcp.notion.com/mcp"
        ]
    },
    "git": {
        "command": "uvx",
        "args": [
        "mcp-server-git"
        ]
    }
    ```

## Design Notes & Conventions

### A Note on the `.vscode` Directory

The choice to store feature artifacts in `.vscode/` is a practical one based on personal habit. Because the `.vscode` folder is ignored in many of our projects, I have been using it to keep personal, project-related files that I don't want to commit.

To override this default, you can specify a different path in your project's `GEMINI.md` file. For example:

> **Feature document root is `.features/` instead of `.vscode/`**

### Architectural Rationale

The design principles of this project emphasize using deterministic tools (`yq`, shell scripts) for procedural tasks, while reserving the LLM for orchestration and analytical functions. This approach has led to the following architectural patterns:

#### LLM Orchestrator with Helper Scripts
This pattern is for complex, interactive commands. The agent acts as an "orchestrator," using `run_shell_command` to execute small helper scripts for predictable steps, while handling the stateful or interactive parts of the workflow itself. This separates procedural tasks (handled by scripts) from analytical tasks (handled by the LLM). `/session:define` is a good example.

#### Subagent Pattern for Focused Tasks
This is the preferred pattern for delegating a complex, one-off task. The main agent instructs the `generalist` sub-agent to execute the task in an isolated session. This provides efficiency and context isolation. Commands like `/session:get-familiar`, `/session:checkpoint`, and `/session:pr` use this pattern.


## Current Limitations and Future Considerations

This workflow is still evolving, and there are some limitations to be aware of:

*   **Requires discipline:** The workflow requires user discipline to be effective.
*   **Semi-automated:** The workflow is not fully automated and requires user guidance.
*   **Actively evolving:** The commands and processes are continuously being refined.

---

## Commands

- `**/session:address-feedback**: Fetches and helps address feedback comments from a GitHub Pull Request.
- `**/session:checkpoint**: Saves a checkpoint of the work done by updating state files using the yq tool.
- `**/session:define**: Starts a conversational session to define a new user story and create its feature directory.
- `**/session:end**: Ends the work session, saving progress and project-wide knowledge to GEMINI.md.
- `**/session:get-familiar`**: Gets familiar with the current code changes by having a subagent generate a summary.
- `**/session:log-research**: Logs a summary of research findings to log.md.
- `**/session:migration**: Migrates an old, single-file feature document to the new directory structure.
- `**/session:new**: Creates a new feature directory from a Shortcut story ID or Notion page URL.
- `**/session:plan**: Analyzes codebase and feature requirements to create an implementation plan.
- `**/session:pr**: Generates a pull request description, creates/updates the PR on GitHub, and saves the link to the feature directory.
- `**/session:review**: Performs a code review of the current branch using a focused sub-agent.
- `**/session:review-devops**: Performs a devops review of the current branch using a focused sub-agent.
- `**/session:review-docs**: Performs a documentation review of the current branch using a focused sub-agent.
- `**/session:start**: Starts a work session by loading context from a feature directory and the project's GEMINI file.
- `**/session:summary**: Generates a human-readable Markdown summary of the entire feature's state.
- `**/session:verify-release**: Verifies a cherry-picked release on the current branch, providing an analysis of any changes found.

---

## Command Details

This section provides a breakdown of individual session commands, their dependencies, and their interactions with the file system and external tools.

### `/session:address-feedback`

-   **Description:** Fetches and helps address unresolved review comments from a feature's GitHub Pull Request.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `read_file`, `pull_request_read`, `run_shell_command`
    -   **External Services:** GitHub
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   GitHub API (to get review comments).
-   **Outputs:**
    -   Appends a summary to `.vscode/<feature-dir>/log.md`.
    -   Modifies project source files to address feedback.

### `/session:checkpoint`

-   **Description:** Saves a snapshot of the work-in-progress by updating the status of tasks and questions and logging a summary of the progress.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Skills:** `yq YAML Processing`
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `run_shell_command`, `generalist`
-   **Inputs:**
    -   The active feature directory path from the conversation.
    -   `.vscode/<feature-dir>/plan.yml`
    -   `.vscode/<feature-dir>/questions.yml`
-   **Outputs:**
    -   Modifies `.vscode/<feature-dir>/plan.yml` in-place.
    -   Modifies `.vscode/<feature-dir>/questions.yml` in-place.
    -   Appends a summary to `.vscode/<feature-dir>/log.md`.

### `/session:define`

-   **Description:** Starts a conversational session to define a new user story and creates the corresponding feature directory.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/create_feature_dir.sh`
    -   **Tools:** `glob`, `grep_search`, `run_shell_command`, `write_file`
-   **Inputs:**
    -   User's command-line arguments (`{{args}}`).
    -   Project source code via `glob` and `grep_search`.
-   **Outputs:**
    -   Creates a new feature directory (e.g., `.vscode/create-user-profile-page/`).
    -   `.vscode/<feature-dir>/description.md` and other placeholder files.
    -   Outputs the "Session Context" block to the chat.

### `/session:end`

-   **Description:** Ends the work session, saving a final summary to the feature directory and persisting any project-wide knowledge to `GEMINI.md`.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Skills:** `yq YAML Processing`
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `read_file`, `run_shell_command`, `replace`, `generalist`
-   **Inputs:**
    -   Session conversation history.
    -   The "Session Context" block from chat history.
    -   `.vscode/<feature-dir>/plan.yml`
    -   `.vscode/<feature-dir>/questions.yml`
-   **Outputs:**
    -   Modifies `.vscode/<feature-dir>/plan.yml` in-place.
    -   Modifies `.vscode/<feature-dir>/questions.yml` in-place.
    -   Appends a final summary to `.vscode/<feature-dir>/log.md`.
    -   Modifies `GEMINI.md` in-place.

### `/session:get-familiar`

-   **Description:** Uses a sub-agent to analyze and summarize the current Git branch's code changes.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `generalist`
-   **Inputs:**
    -   Local Git repository state.
-   **Outputs:**
    -   Writes a summary of code changes to standard output.

### `/session:log-research`

-   **Description:** Logs a summary of research findings to the feature's `log.md` file.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `generalist`
-   **Inputs:**
    -   Session conversation history.
-   **Outputs:**
    -   Delegates appending a timestamped report to `.vscode/<feature-dir>/log.md` to a sub-agent.

### `/session:migration`

-   **Description:** Migrates a legacy, single-file feature markdown document into the multi-file directory structure.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Scripts:** `scripts/migrate_feature_file.sh`
    -   **Tools:** `run_shell_command`
-   **Inputs:**
    -   A single markdown file path provided as an argument.
-   **Outputs:**
    -   Creates a new directory named after the input file.
    -   Populates the new directory with `description.md`, `plan.yml`, `questions.yml`, etc.
    -   Archives the original file by renaming it with a `.migrated` extension.

### `/session:new`

-   **Description:** Creates a feature directory from a Shortcut story ID or Notion page URL, fetching resources to populate the `description.md` file.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/create_feature_dir.sh`
    -   **Tools:** `run_shell_command`, `generalist`, `read_file`
    -   **External Services:** Shortcut, Notion, GitHub
-   **Inputs:**
    -   Delegates API calls to Shortcut, Notion, and GitHub to a sub-agent.
    -   Reads `GEMINI.md`.
-   **Outputs:**
    -   Creates a new directory and placeholder files.
    -   Delegates writing the `description.md` to a sub-agent.
    -   Outputs the "Session Context" block to the chat.

### `/session:plan`

-   **Description:** Analyzes codebase and feature requirements to create an implementation plan.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Tools:** `read_file`, `glob`, `grep_search`, `write_file`
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Codebase files via `glob` and `grep_search`.
    -   User input during interactive planning.
-   **Outputs:**
    -   Generates and writes initial `.vscode/<feature-dir>/plan.yml`
    -   Generates and writes initial `.vscode/<feature-dir>/questions.yml`

### `/session:pr`

-   **Description:** Generates a pull request description, creates or updates the PR on GitHub, and saves the PR link to the feature directory.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `run_shell_command`, `search_pull_requests`, `read_file`, `create_pull_request`, `update_pull_request`, `write_file`, `ask_user`, `generalist`
    -   **External Services:** GitHub
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Git repository state.
    -   `.git/pull_request_template.md`
    -   Feature directory files (`plan.yml`, `log.md`).
-   **Outputs:**
    -   Delegates PR description generation to a sub-agent.
    -   Creates or updates a pull request on GitHub.
    -   Writes the PR link to `.vscode/<feature-dir>/description.md`.
    -   Outputs an updated "Session Context" block to the chat.
    -   Saves `pull_request_descr.md` to `.vscode/<feature-dir>/` (as a fallback).

### `/session:review`

-   **Description:** Performs a code review of the current branch using a focused sub-agent.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `run_shell_command`, `read_file`, `generalist`
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Git repository state.
    -   Reads back the `.vscode/<feature-dir>/review.yml` for verification.
-   **Outputs:**
    -   Delegates writing `.vscode/<feature-dir>/review.yml` to a sub-agent.

### `/session:review-devops`

-   **Description:** Performs a devops review of the current branch using a focused sub-agent.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `run_shell_command`, `read_file`, `generalist`
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Git repository state.
    -   Reads back the `.vscode/<feature-dir>/devops-review.yml` for verification.
-   **Outputs:**
    -   Delegates writing `.vscode/<feature-dir>/devops-review.yml` to a sub-agent.

### `/session:review-docs`

-   **Description:** Performs a documentation review of the current branch using a focused sub-agent.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `run_shell_command`, `read_file`, `generalist`
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Git repository state.
    -   Reads back the `.vscode/<feature-dir>/docs-review.yml` for verification.
-   **Outputs:**
    -   Delegates writing `.vscode/<feature-dir>/docs-review.yml` to a sub-agent.

### `/session:start`

-   **Description:** Starts a work session by loading context from a feature directory and the project's `GEMINI.md` file.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Scripts:** `scripts/load_context_files.sh`
    -   **Tools:** `run_shell_command`
-   **Inputs:**
    -   The output of the `load_context_files.sh` script, which contains the content of all files in the feature directory and `GEMINI.md`.
-   **Outputs:**
    -   Outputs the "Session Context" block to the chat.

### `/session:summary`

-   **Description:** Generates a human-readable Markdown summary of the feature's entire state.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Scripts:** `scripts/load_context_files.sh`
    -   **Tools:** `run_shell_command`, `write_file`
-   **Inputs:**
    -   The output of the `load_context_files.sh` script.
-   **Outputs:**
    -   `.vscode/<feature-dir>/_SUMMARY.md`

### `/session:verify-release`

-   **Description:** Verifies a cherry-picked release branch against its original commits and provides an analysis of any discrepancies.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Scripts:** `scripts/verify-release.sh`
    -   **Tools:** `run_shell_command`
-   **Inputs:**
    -   Local Git repository (branches, commits, commit messages, and patch data).
-   **Outputs:**
    -   Writes temporary patch files for comparison.
    -   Writes a verification report to standard output.
