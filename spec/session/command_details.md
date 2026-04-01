# Command Details

This section provides a breakdown of individual session commands, their dependencies, and their interactions with the file system and external tools.

## `/session:address-feedback`

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
    -   Appends a summary to `.features/<feature-dir>/log.md`.
    -   Modifies project source files to address feedback.

## `/session:checkpoint`

-   **Description:** Saves a snapshot of the work-in-progress by updating the status of tasks and questions and logging a summary of the progress.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `run_shell_command`
-   **Inputs:**
    -   The active feature directory path from the conversation.
    -   `.features/<feature-dir>/plan.yml`
    -   `.features/<feature-dir>/questions.yml`
-   **Outputs:**
    -   Modifies `.features/<feature-dir>/plan.yml` in-place.
    -   Modifies `.features/<feature-dir>/questions.yml` in-place.
    -   Appends a summary to `.features/<feature-dir>/log.md`.

## `/session:define "{USER STORY DESCRIPTION}"`

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
    -   Creates a new feature directory (e.g., `.features/create-user-profile-page/`).
    -   `.features/<feature-dir>/description.md` and other placeholder files.
    -   Outputs the "Session Context" block to the chat.

## `/session:end`

-   **Description:** Ends the work session, saving a final summary to the feature directory and persisting any project-wide knowledge to `AGENTS.md`.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Skills:** None
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `run_shell_command`, `replace`
-   **Inputs:**
    -   Session conversation history.
    -   The "Session Context" block from chat history.
    -   `.features/<feature-dir>/plan.yml`
    -   `.features/<feature-dir>/questions.yml`
-   **Outputs:**
    -   Modifies `.features/<feature-dir>/plan.yml` in-place.
    -   Modifies `.features/<feature-dir>/questions.yml` in-place.
    -   Appends a final summary to `.features/<feature-dir>/log.md`.
    -   Modifies `AGENTS.md` in-place.

## `/session:get-familiar`

-   **Description:** Uses a sub-agent to analyze and summarize the current Git branch's code changes.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh`
    -   **Tools:** `generalist`
-   **Inputs:**
    -   Local Git repository state.
-   **Outputs:**
    -   Writes a summary of code changes to standard output.

## `/session:log-research`

-   **Description:** Logs a summary of research findings to the feature's `log.md` file.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/append_to_log.sh`
    -   **Tools:** `generalist`
-   **Inputs:**
    -   Session conversation history.
-   **Outputs:**
    -   Delegates appending a timestamped report to `.features/<feature-dir>/log.md` to a sub-agent.

## `/session:migration`

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

## `/session:new {SHORTCUT ID}`

-   **Description:** Creates a feature directory from a Shortcut story ID or Notion page URL, fetching resources to populate the `description.md` file.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/create_feature_dir.sh`
    -   **Tools:** `run_shell_command`, `generalist`, `read_file`
    -   **External Services:** Shortcut, Notion, GitHub
-   **Inputs:**
    -   Delegates API calls to Shortcut, Notion, and GitHub to a sub-agent.
    -   Reads `AGENTS.md`.
-   **Outputs:**
    -   Creates a new directory and placeholder files.
    -   Delegates writing the `description.md` to a sub-agent.
    -   Outputs the "Session Context" block to the chat.

## `/session:plan`

-   **Description:** Analyzes codebase and feature requirements to create an implementation plan.
-   **Orchestration Pattern:** LLM Orchestrator
-   **Dependencies:**
    -   **Tools:** `read_file`, `glob`, `grep_search`, `write_file`
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Codebase files via `glob` and `grep_search`.
    -   User input during interactive planning.
-   **Outputs:**
    -   Generates and writes initial `.features/<feature-dir>/plan.yml`
    -   Generates and writes initial `.features/<feature-dir>/questions.yml`

## `/session:pr`

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
    -   Writes the PR link to `.features/<feature-dir>/description.md`.
    -   Outputs an updated "Session Context" block to the chat.
    -   Saves `pull_request_descr.md` to `.features/<feature-dir>/` (as a fallback).

## `/session:review`

-   **Description:** Performs a code review of the current branch using a focused sub-agent.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh` (called by the sub-agent, not the main session)
    -   **Tools:** `read_file`, `generalist`
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Git repository state (fetched by the sub-agent).
    -   Reads back the `.features/<feature-dir>/review.yml` for verification.
-   **Outputs:**
    -   Delegates writing `.features/<feature-dir>/review.yml` to a sub-agent.

## `/session:review-devops`

-   **Description:** Performs a devops review of the current branch using a focused sub-agent.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh` (called by the sub-agent, not the main session)
    -   **Tools:** `read_file`, `generalist`
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Git repository state (fetched by the sub-agent).
    -   Reads back the `.features/<feature-dir>/devops-review.yml` for verification.
-   **Outputs:**
    -   Delegates writing `.features/<feature-dir>/devops-review.yml` to a sub-agent.

## `/session:review-docs`

-   **Description:** Performs a documentation review of the current branch using a focused sub-agent.
-   **Orchestration Pattern:** Subagent Pattern
-   **Dependencies:**
    -   **Scripts:** `scripts/get_git_context.sh` (called by the sub-agent, not the main session)
    -   **Tools:** `read_file`, `generalist`
-   **Inputs:**
    -   The "Session Context" block from chat history.
    -   Git repository state (fetched by the sub-agent).
    -   Reads back the `.features/<feature-dir>/docs-review.yml` for verification.
-   **Outputs:**
    -   Delegates writing `.features/<feature-dir>/docs-review.yml` to a sub-agent.

## `/session:start {FEATURE DIRECTORY NAME}`

-   **Description:** Starts a work session by loading context from a feature directory and the project's `AGENTS.md` file.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Scripts:** `scripts/load_context_files.sh`
    -   **Tools:** `run_shell_command`
-   **Inputs:**
    -   The output of the `load_context_files.sh` script, which contains the content of all files in the feature directory and `AGENTS.md`.
-   **Outputs:**
    -   Outputs the "Session Context" block to the chat.

## `/session:summary`

-   **Description:** Generates a human-readable Markdown summary of the entire feature's state.
-   **Orchestration Pattern:** LLM Orchestrator with Helper Scripts
-   **Dependencies:**
    -   **Scripts:** `scripts/load_context_files.sh`
    -   **Tools:** `run_shell_command`, `write_file`
-   **Inputs:**
    -   The output of the `load_context_files.sh` script.
-   **Outputs:**
    -   `.features/<feature-dir>/_SUMMARY.md`

## `/session:verify-release`

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
