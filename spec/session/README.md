# A Session-Based Workflow for Development

> **Note on Active Development:** This project is under active development and is used as a playground for exploring and experimenting with AI-assisted development concepts. As such, commands and workflows may change or break unexpectedly.

> **Important Note on Feature Directories:** By default, feature directories are located in `.vscode/`. However, this can be overridden by a custom path defined in your project's `AGENTS.md`. For more details, refer to [A Note on the `.vscode` Directory](#a-note-on-the-vscode-directory).

## Overview

This document describes a suite of AI assistant commands for a structured, session-based workflow, compatible with both Gemini CLI and Claude Code. The goal is to provide a semi-automated development loop that manages context by separating two types of knowledge:

1.  **Feature-Specific Knowledge:** The details, requirements, implementation plan, and progress related to a single user story. This context is ephemeral and only relevant for the duration of the task.
2.  **Project-Wide Knowledge:** Architectural patterns, conventions, and learnings that should persist and be shared across all future development in the repository.

This suite of commands orchestrates the flow of information between the user, the codebase, external tools, and dedicated documents that store these two types of knowledge. This creates persistent storage for both task-specific and project-wide information.


## How to Invoke Commands

Commands follow a `/namespace:command` syntax and are typed directly in your AI
assistant's chat interface — not in the terminal.

- **Gemini CLI:** type `/session:start sc-1234` in the Gemini chat prompt
- **Claude Code:** type `/session:start sc-1234` in the Claude Code chat prompt

Both tools use the same `/session:` prefix. Arguments are passed after the command name.

## Getting Started: Session Entry Points

To begin a work session, there are three primary commands:

*   **`/session:define`**: Use this to define a **new user story from scratch**. This command guides you through a conversational process to capture requirements and creates a new feature directory.
    ```bash
    /session:define "define user story in natural language, you can link @filenames as well"
    ```

*   **`/session:new`**: Use this to **create a feature document from an existing user story ID or Notion page URL**. This command fetches information from external services (like Shortcut or Notion) to pre-populate your feature directory.
    ```bash
    # Creates feature document fetching user story sc-1234 from Shortcut
    /session:new sc-1234

    # Creates feature document fetching user story from Notion
    /session:new https://notion_page_url.com
    ```

*   **`/session:start`**: Use this to **resume work on an existing feature**. This command loads context from a previously created feature directory into your current session.
    ```bash
    # Open feature directory in default path .vscode/sc-1234
    /session:start sc-1234
    
    # Open feature directory located in a custom path
    /session:start .features/sc-1234
    ```

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

-   **`yq` v4+:** Used for modifying `.yml` state files (`brew install yq`). Make sure
    you install [mikefarah/yq](https://github.com/mikefarah/yq), not the Python-based
    `yq` — they have different syntax.
-   **Node.js / `npx`:** Required for several MCP servers.
-   **`uv` / `uvx`:** Required for the Git MCP server (`brew install uv`).
-   **Gemini skills (Gemini CLI only):** One skill needs to be installed locally.
    From the repo root, run:
    ```bash
    gemini skills install ~/.ai-session/gemini/tdd-skill
    ```
-   **MCP Servers:** Integrations with Shortcut, Notion, Git, and GitHub are used by
    several commands. Configure them in your tool's settings file.

    **Gemini CLI** — add to `~/.gemini/settings.json` under `"mcpServers"`:
    ```json
    "shortcut": {
        "command": "npx",
        "args": ["-y", "@shortcut/mcp@latest"],
        "env": {
            "SHORTCUT_API_TOKEN": "your_shortcut_api_token_here"
        }
    },
    "notion": {
        "command": "npx",
        "args": ["-y", "mcp-remote", "https://mcp.notion.com/mcp"]
        // Note: requires browser-based OAuth on first run
    },
    "git": {
        "command": "uvx",
        "args": ["mcp-server-git"]
    },
    "github": {
        "command": "npx",
        "args": ["-y", "@github/mcp-server"],
        "env": {
            "GITHUB_PERSONAL_ACCESS_TOKEN": "your_github_token_here"
        }
    }
    ```

    **Claude Code** — add the same block to `~/.claude/settings.json` under
    `"mcpServers"`. The format is identical.

    For more details see the docs for
    [Shortcut](https://www.shortcut.com/blog/why-we-built-the-shortcut-mcp-server),
    [Notion](https://developers.notion.com/guides/mcp/get-started-with-mcp),
    [Git](https://pypi.org/project/mcp-server-git/), and
    [GitHub](https://github.com/github/github-mcp-server).

## Commands

- **/session:address-feedback**: Fetches and helps address feedback comments from a GitHub Pull Request.
- **/session:checkpoint**: Saves a checkpoint of the work done by updating state files.
- **/session:define {USER STORY DESCRIPTION}**: Starts a conversational session to define a new user story and create its feature directory.
- **/session:end**: Ends the work session, saving progress and project-wide knowledge to AGENTS.md.
- **/session:get-familiar**: Gets familiar with the current code changes by having a subagent generate a summary.
- **/session:log-research**: Logs a summary of research findings to log.md.
- **/session:migration**: Migrates an old, single-file feature document to the new directory structure.
- **/session:new {SHORTCUT ID}**: Creates a new feature directory from a Shortcut story ID or Notion page URL.
- **/session:plan**: Analyzes codebase and feature requirements to create an implementation plan.
- **/session:pr**: Generates a pull request description, creates/updates the PR on GitHub, and saves the link to the feature directory.
- **/session:review**: Performs a code review of the current branch using a focused sub-agent.
- **/session:review-devops**: Performs a devops review of the current branch using a focused sub-agent.
- **/session:review-docs**: Performs a documentation review of the current branch using a focused sub-agent.
- **/session:start {FEATURE DIRECTORY NAME}**: Starts a work session by loading context from a feature directory and the project's AGENTS.md file.
- **/session:summary**: Generates a human-readable Markdown summary of the entire feature's state.
- **/session:verify-release**: Verifies a cherry-picked release on the current branch, providing an analysis of any changes found.

Please check the file [Command Details](command_details.md) for a full breakdown of each command's dependencies, inputs, and outputs.

## Core Concepts

-  **Session:** A session represents the working context for a single feature or user story.

    In this workflow, a session is not only tied to the terminal or chat history. Instead, it is primarily defined by the feature directory, which stores the description, implementation plan, open questions, progress log, and review notes for the feature.

    The terminal session is still used during development, but the feature directory acts as a more stable source of context. This makes it easier to resume work across multiple days and allows the LLM to use structured information rather than incomplete conversational history.
-   **Feature Directory:** A directory located in `.vscode/` (e.g., `.vscode/sc-12345/`). It contains a mix of Markdown files (like `description.md`, `log.md`) and structured YAML files (`plan.yml`, `questions.yml`, `review.yml`) that hold the state for a specific feature. This serves as the "session memory." See [`spec/session/example-feature-document/`](example-feature-document/) in this repo for a complete example.

    Example:
    ```
    .vscode/sc-12345/
        description.md
        plan.yml
        questions.yml
        log.md
        review.yml
    ```
-   **Project Document (`AGENTS.md`):** A file at the root of **each of your own
    repositories** (not this repo) that stores project-wide context, architectural
    guidelines, and conventions. The AI reads this file to understand the project it
    is helping with. This serves as the "project memory" and helps maintain consistency
    across all sessions on that project.

    Examples of knowledge to include in `AGENTS.md`:
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

## Design Notes & Conventions

### A Note on the `.vscode` Directory

The choice to store feature artifacts in `.vscode/` is a practical one based on personal habit. Because the `.vscode` folder is ignored in many of our projects, I have been using it to keep personal, project-related files that I don't want to commit.

To override this default, you can specify a different path in your project's `AGENTS.md` file. For example:

> **Feature directories root is `.features/` instead of `.vscode/`**

### Architectural Rationale

The design principles of this project emphasize using deterministic tools (`yq`, shell scripts) for procedural tasks, while reserving the LLM for orchestration and analytical functions. This approach has led to the following architectural patterns:

#### LLM Orchestrator with Helper Scripts
This pattern is for complex, interactive commands. The agent acts as an "orchestrator," using `run_shell_command` to execute small helper scripts for predictable steps, while handling the stateful or interactive parts of the workflow itself. This separates procedural tasks (handled by scripts) from analytical tasks (handled by the LLM). `/session:define` is a good example.

#### Subagent Pattern for Focused Tasks
This is the preferred pattern for delegating a complex, one-off task. The main agent instructs the `generalist` sub-agent to execute the task in an isolated session. This provides efficiency and context isolation. Commands like `/session:get-familiar`, `/session:checkpoint`, and `/session:pr` use this pattern.


## Current Limitations and Future Considerations

This workflow is still evolving, and there are some limitations to be aware of:

*   **Tokens usage:** Because of the nature of constantly loading the context from external files, the commands have a high tokens consumption.
*   **Requires discipline:** The workflow requires user discipline to be effective.
*   **Semi-automated:** The workflow is not fully automated and requires user guidance.
*   **Actively evolving:** The commands and processes are continuously being refined.