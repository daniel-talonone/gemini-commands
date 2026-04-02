# Generated from claude/session/new.md — do not edit directly.
# Run scripts/gen_headless.sh to regenerate.

You are an assistant who bootstraps feature development by creating a feature directory and a placeholder description. The user has provided an identifier: `{{args}}`.

**Orchestration Actions:**

1.  **Determine Identifier Type and Feature Name:**
    *   A `feature_name` variable must be determined from `{{args}}`. Use `run_shell_command` to execute logic to set this variable.
    *   If `{{args}}` starts with "sc-", the `feature_name` is `{{args}}`.
    *   If `{{args}}` is a URL containing "notion.so", the `feature_name` must be derived from the last part of the URL path (the slug), for example by using `basename`.

2.  **Scaffold Directory:**
    *   Call the `create_feature_dir.sh` helper script using `run_shell_command` to create the directory and all placeholder files. The command is: `$AI_SESSION_HOME/scripts/create_feature_dir.sh "$($AI_SESSION_HOME/scripts/resolve_feature_dir.sh "YOUR_DERIVED_FEATURE_NAME")"`.
    *   The full path to the created directory, from `$AI_SESSION_HOME/scripts/resolve_feature_dir.sh`, must be stored for subsequent steps.

3.  **Synthesize Placeholder Description:**
    *   Direct context gathering from the identifier is not possible in this headless environment as it requires specialized interactive tools. Therefore, you must create a placeholder `description.md` file instead.
    *   Use `run_shell_command` with `printf` to write the following content to the `description.md` file in the directory created above. The `{{...}}` placeholders must be replaced with the actual values.

    ```markdown
    # {{feature_name}}

    **Source:** {{args}}

    **TODO: Content Synthesis Skipped**

    In an interactive session, this file would be populated by fetching and synthesizing content from the source identifier and any linked resources. This step was skipped because the required data-fetching tools are not available in this headless environment.
    ```

4.  **Establish Session Context (Final Step):**
    *   Read the content of the `description.md` file you just created using `run_shell_command` with `cat`.
    *   Read the content of `AGENTS.md` from the project root (fall back to `GEMINI.md` if not present) using `run_shell_command`.
    *   Format and display using the following Markdown structure EXACTLY, replacing the placeholders with the real values:

        ```markdown
        ### ✨ Session Context Loaded for `{{feature_name}}`

        **Description:**
        > {{The synthesized description from the new description.md}}

        This context is now available for all subsequent commands.
        ```
    *   After printing this block, the command is complete.
