---
description: Generates a human-readable Markdown summary of the feature's state.
---

You are a reporting assistant. Your goal is to create a comprehensive Markdown summary of the current feature's state.

The user has provided a feature directory name as an argument: `$ARGUMENTS`.

**Process:**

1.  **Load Context:**
    *   Execute the `load_context_files.sh` script using the Bash tool to gather all the context from the feature directory `.vscode/$ARGUMENTS`.
    *   Example: `$AI_SESSION_HOME/scripts/load_context_files.sh ".vscode/$ARGUMENTS"`
    *   The script's output is a single string containing the content of all files (`description.md`, `plan.yml`, `questions.yml`, etc.) each preceded by `--- FILE: <filename> ---`.

2.  **Synthesize Markdown Report:**
    *   Construct a single Markdown string that consolidates all the information from the script's output.
    *   **CRITICAL:** Convert the structured YAML data into human-readable Markdown:
        *   For `plan.yml`, create a checklist. An item with `status: 'done'` → `- [x] Task description`. Any other status → `- [ ] Task description`.
        *   For `questions.yml`, create a Q&A list showing the question, its status, and the answer if it exists.
        *   For `review.yml`, create a list of feedback items.
    *   Organize the final document with clear headings for each section (e.g., `## Plan`, `## Open Questions`, `## Work Log`).

3.  **Write Summary File:**
    *   Use the Write tool to save the complete Markdown string to `_SUMMARY.md` inside the `.vscode/$ARGUMENTS` directory.
    *   **This command must always overwrite the file if it exists.**

4.  **Confirm:** Announce that the summary file has been created/updated for `$ARGUMENTS`.
