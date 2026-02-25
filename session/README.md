# Session Workflow Commands

This document describes a suite of custom Gemini CLI commands designed to implement a structured, session-based workflow for software development. The workflow is centered around a "feature document" that acts as a single source of truth for a task.

## Project Philosophy

The primary goal of this command suite is to create a robust, semi-automated development loop that intelligently manages context. The core idea is to differentiate between two types of knowledge:

1.  **Feature-Specific Knowledge:** The details, requirements, implementation plan, and progress related to a single user story. This context is ephemeral and only relevant for the duration of the task.
2.  **Project-Wide Knowledge:** Enduring architectural patterns, conventions, and learnings that should persist and be shared across all future development in the repository.

This suite of commands orchestrates the flow of information between the user, the codebase, external tools, and dedicated documents that store these two types of knowledge, creating a persistent "memory" for both the specific task and the overall project.

## Core Concepts

-   **Feature Document:** A Markdown file located in the `.vscode/` directory (e.g., `.vscode/sc-12345.md`). It contains the requirements, implementation plan, and progress log for a specific feature or bug fix. This serves as the "session memory."
-   **Project Document (`GEMINI.md`):** A global file that stores project-wide context, architectural guidelines, and conventions. This serves as the "project memory."

## Dependencies

These commands rely on integrations with several external services (MCPs). The prompts are written to leverage the specific tools available for each:

-   **[Shortcut](https://shortcut.com/):** Used for fetching story details to create and contextualize feature documents.
-   **[Notion](https://www.notion.so/):** Used to pull in additional context from linked documentation.
-   **[Git](https://git-scm.com/):** Used for managing branches, viewing diffs, and understanding the local state of the code.
-   **[GitHub](https://github.com/google/gemini-cli):** Used for creating and managing pull requests and interacting with code reviews. The link points to the core Gemini CLI repository, which provides the GitHub integration.

## Design Notes & Conventions

### A Note on the `.vscode` Directory

The choice to store feature documents in `.vscode/` is a practical one based on personal habit. Because the `.vscode` folder is ignored in many of our projects, I have been using it to keep personal, project-related files that I don't want to commit.

For that reason, it was natural for me to store the feature documents in this directory when I started this project. I recognize this is not an ideal solution and can clearly be improved, but it was very practical for getting started.

---

## Commands

Here is a detailed description of each command in the session suite.

### /session:new `[story-id]`

**Description:** Creates a new feature document from a Shortcut story ID. It fetches the story's name, description, and comments, and also pulls in content from any linked Notion pages or other Shortcut stories. This file becomes the central reference for the task.

**Usage:**
`bash
/session:new sc-12345
`

### /session:start `[feature-doc]`

**Description:** Starts a work session. This command loads the context from the specified feature document (e.g., `sc-12345.md`) and the main `GEMINI.md` project file, preparing the assistant for the work ahead.

**Usage:**
`bash
/session:start sc-12345.md
`

### /session:plan

**Description:** Analyzes the codebase and the loaded feature document to generate a detailed, step-by-step implementation plan. It populates the "Next Steps" and "Open Questions" sections of the feature document, providing a clear path forward. This is a planning phase only.

**Usage:**
`bash
/session:plan
`

### /session:checkpoint

**Description:** Saves a timestamped log of the work done during a session. It reviews the conversation history, records completed tasks, notes decisions made, and updates the "Work Log," "Next Steps," and "Open Questions" sections in the feature document. This ensures progress is not lost between sessions.

**Usage:**
`bash
/session:checkpoint
`

### /session:review

**Description:** Performs a critical, context-aware code review of the current git branch. It uses the active feature document as the source of truth for acceptance criteria and saves its feedback (a numbered list of required changes) to a "## Code Review Feedback" section within that same document.

**Usage:**
`bash
/session:review
`

### /session:review_from_branch

**Description:** Performs a code review similar to `/session:review`, but derives its context directly from the Shortcut story ID found in the current git branch name (e.g., `sc-12345-my-feature`). It posts the feedback as a comment on the corresponding Shortcut story.

**Usage:**
`bash
/session:review_from_branch
`

### /session:pr

**Description:** Generates a comprehensive pull request description. It uses the active feature document, a template from `.vscode/pull_request_template.md`, and the `git diff` of the current branch. It can then update an existing PR on GitHub, or, with user approval, create a new one. It also saves the PR link to the feature document for future reference.

**Usage:**
`bash
/session:pr
`

### /session:pr_from_branch

**Description:** Generates a pull request description, but like `review_from_branch`, it uses the Shortcut story linked in the git branch name for context instead of a feature document. It then updates the corresponding PR on GitHub or creates a local file.

**Usage:**
`bash
/session:pr_from_branch
`

### /session:address-feedback

**Description:** Fetches unresolved review comments from the feature's linked GitHub Pull Request. It then guides the user through addressing each comment one by one, proposing a plan, and implementing the changes upon approval, while documenting the rationale for each change.

**Usage:**
`bash
/session:address-feedback
`

### /session:end

**Description:** Finalizes a work session. It performs one last checkpoint to save all remaining progress to the feature document. Crucially, it also analyzes the entire session to identify any new, project-wide knowledge (like new architectural patterns or conventions) and saves it to the global `GEMINI.md` file for future reference.

**Usage:**
`bash
/session:end
`

---

## Typical Workflows

### Full Session-Based Workflow

This is the most common and structured way to use the command suite.

1.  **Initialize the feature:**
    `bash
    /session:new sc-12345
    `

2.  **Start your work session:**
    `bash
    /session:start sc-12345.md
    `

3.  **Create an implementation plan:**
    `bash
    /session:plan
    `

4.  **(Work on the code using Gemini's assistance...)**

5.  **Log your progress periodically:**
    `bash
    /session:checkpoint
    `

6.  **(Finish the implementation...)**

7.  **Review your changes against the requirements:**
    `bash
    /session:review
    `

8.  **Prepare and create the pull request:**
    `bash
    /session:pr
    `

9.  **Address PR feedback from teammates:**
    `bash
    /session:address-feedback
    `

10. **End the session and save long-term knowledge:**
    `bash
    /session:end
    `

### Quick Actions from a Branch

These commands are useful for quickly performing a single action without a full session context.

**Scenario 1: Code Review from a Branch**

1.  Checkout a feature branch (e.g., `sc-54321-fix-login-bug`).
2.  Run a review based on the linked Shortcut story:
    `bash
    /session:review_from_branch
    `

**Scenario 2: PR Description from a Branch**

1.  Checkout the feature branch and ensure your changes are complete.
2.  Generate the PR description from the linked Shortcut story:
    `bash
    /session:pr_from_branch
    `
## Future Considerations

For developers looking to extend or improve this suite, here are some potential ideas:

-   **Configurable Artifact Path:** The path for feature documents is currently hardcoded to `.vscode/`. A future improvement could be to make this path configurable, perhaps via a setting in the global `GEMINI.md` file.
-   **Enhanced Error Handling:** The commands could benefit from more robust error handling, providing clearer feedback when a dependency (like a Shortcut story or Notion page) is not found.
-   **Abstracting Core Logic:** The core logic for interacting with feature documents (reading, updating sections) could be abstracted into a shared skill or prompt to reduce duplication across commands.
-   **Interactive Setup:** A command like `/session:init` could be created to help set up the necessary `GEMINI.md` and `.vscode/pull_request_template.md` files for a new project.
