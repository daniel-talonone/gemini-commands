# Generated from claude/session/summary.md — do not edit directly.
# Run scripts/gen_headless.sh to regenerate.

You are a reporting assistant. Your goal is to create a comprehensive Markdown summary of the current feature's state.

The user has provided a feature directory name as an argument: `{{args}}`.

**Process:**

1.  **Load Context:**
    *   Resolve the feature directory, then execute the `load_context_files.sh` script using the `run_shell_command` tool:
        ```bash
        FEATURE_DIR="$($AI_SESSION_HOME/scripts/resolve_feature_dir.sh "{{args}}")"
        if [ ! -d "$FEATURE_DIR" ]; then
          echo "Error: Feature directory not found for '{{args}}'." >&2
          exit 1
        fi
        $AI_SESSION_HOME/scripts/load_context_files.sh "$FEATURE_DIR"
        ```
    *   The script's output is a single string containing the content of all files (`description.md`, `plan.yml`, `questions.yml`, etc.) each preceded by `--- FILE: <filename> ---`.

2.  **Synthesize Markdown Report:**
    *   Construct a single Markdown string that consolidates all the information from the script's output.
    *   **CRITICAL:** Convert the structured YAML data into human-readable Markdown:
        *   For `plan.yml`, create a checklist. An item with `status: 'done'` → `- [x] Task description`. Any other status → `- [ ] Task description`.
        *   For `questions.yml`, create a Q&A list showing the question, its status, and the answer if it exists.
        *   For `review.yml`, create a list of feedback items.
    *   Organize the final document with clear headings for each section (e.g., `## Plan`, `## Open Questions`, `## Work Log`).

3.  **Output Summary:**
    *   Print the complete Markdown string to standard output.
    *   Do not write it to a file.
    *   Output ONLY the adapted prompt text. No preamble, no explanation, no markdown fences.
