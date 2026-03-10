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

#### From Single File to Feature Directory
This command suite originally used a single Markdown file. This proved brittle as it relied on the `replace` tool, which often failed. The first refactoring split this file into a **feature directory**, enabling safer, more reliable file-specific operations.

#### From Markdown Lists to Structured YAML
The next improvement was to change stateful lists (plans, questions, etc.) from Markdown to **structured YAML**. This allowed the agent to programmatically parse and modify the data in memory before writing it back, which was more robust than text manipulation.

#### From In-Memory Parsing to `yq`
The final evolution was to offload all YAML manipulation to the `yq` command-line tool. By activating a `yq-skill` and using `run_shell_command`, the agent can issue direct, atomic commands (e.g., `yq -i '.field = "value"' file.yml`). This is the most robust and reliable pattern, as it uses a specialized, deterministic tool for all state updates.

#### From LLM Procedures to Helper Scripts
Following the same philosophy, any procedural, deterministic logic (especially file system operations) is being migrated from the LLM's prompt into dedicated helper scripts located in the `scripts/` directory. LLMs are non-deterministic and can be unreliable when asked to follow a strict sequence of procedural steps. Encapsulating these steps in a script makes the commands more robust.

The LLM's role is shifted from *performing* the steps to *executing* the script that performs them. These scripts are called using a portable, reliable execution pattern that leverages the known conventional path for global commands (`$HOME/.gemini/commands`). This refactoring effort is ongoing.

Examples include:
- The `/session:new` command, which uses `scripts/create_feature_dir.sh` to scaffold the feature directory.
- The `/session:checkpoint` command, which uses `scripts/append_to_log.sh` to add a timestamped entry to the log file, ensuring consistent formatting and avoiding file corruption.

#### From Helper Scripts to Hybrid Orchestrators
The latest and most powerful evolution of this architecture is the **hybrid orchestrator** pattern. This pattern resolves a key limitation: a command's `prompt` can either be a non-interactive shell script OR a flexible LLM prompt, but not both. The hybrid model provides the best of both worlds.

The pattern is as follows:
1.  The command's `prompt` is defined as a **high-level LLM prompt**, not a shell script. This prompt acts as an "orchestrator" for the entire command workflow.
2.  This orchestrator agent uses the `run_shell_command` tool to execute small, deterministic, single-purpose **helper scripts** for predictable steps where precision is critical (e.g., generating a filename with a specific timestamp format).
3.  The orchestrator agent then handles the complex, stateful, or interactive parts of the workflow itself. This can include loops, conditional logic, and calling other tools like `read_file` or `replace`.

This architecture allows a single, self-contained command to have a complex, interactive, AI-powered workflow. The `/session:prepare-release` command is the canonical example of this pattern. It uses a helper script to reliably create a release branch, then the main orchestrator agent manages the complex loop of cherry-picking commits and handling potential merge conflicts by analyzing them and asking the user for approval on proposed fixes. This balances the reliability of scripts for deterministic tasks with the analytical flexibility of the LLM for complex ones.

---

## Commands

(Content of all command descriptions remains the same)
...
