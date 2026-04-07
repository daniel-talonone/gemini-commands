# Generated from claude/session/new.md — do not edit directly.
# Run scripts/gen_headless.sh to regenerate.

You are an assistant who bootstraps feature development. An identifier is provided via `{{args}}`.

**Orchestration Actions:**

1.  **Determine Identifier Type and Feature Name:**
    *   If `{{args}}` starts with `sc-`, the `feature_name` is `{{args}}`.
    *   If `{{args}}` is a URL containing `notion.so`, the `feature_name` is derived from the last part of the URL path (the slug). For example, from `https://www.notion.so/t1rnd/My-Page-Title-a1b2c3d4` the name would be `My-Page-Title-a1b2c3d4`.
    *   You will use this `feature_name` in subsequent steps.

2.  **Scaffold Directory:**
    *   Execute the `create_feature_dir.sh` helper script using `run_shell_command` to create the directory and all placeholder files.
    *   The command is: `$AI_SESSION_HOME/scripts/create_feature_dir.sh "$($AI_SESSION_HOME/scripts/resolve_feature_dir.sh "{{feature_name}}")"`. Use the `feature_name` you determined in step 1. The shell will expand the `$AI_SESSION_HOME` variable.

3.  **Gather Context and Synthesize Description:**
    *   The target file for the description is `"$($AI_SESSION_HOME/scripts/resolve_feature_dir.sh "{{feature_name}}")/description.md"`. Use the `feature_name` from step 1.

    a.  **Fetch Primary Content:**
        *   If the identifier is a Shortcut story ID (e.g., "sc-12345"), use the Shortcut MCP to fetch the story.
        *   If the identifier is a Notion URL, use the Notion MCP to fetch the page.

    b.  **Extract Key Information & Find Linked Resources:**
        *   From the fetched content, extract the title, description, comments, and canonical URL.
        *   Scan the description and comments for any Markdown links or raw URLs. Create a list of these linked URLs.

    c.  **Fetch Linked Resources:**
        *   For each linked URL:
            *   If it's a Shortcut link, use the Shortcut MCP.
            *   If it's a Notion link, use the Notion MCP.
            *   If it's a GitHub link to a file, use the GitHub MCP to fetch file contents.
            *   If it's any other HTTP link, use the WebFetch tool.
        *   Limit recursion to 1 level deep.

    d.  **Synthesize the Final Description:**
        *   Combine all gathered information into a single, comprehensive Markdown document.
        *   At the very top, add a "Source:" line with the original identifier `{{args}}`.
        *   Add the title as a main heading.
        *   Include the full description and comments.
        *   Create a "## Linked Resources" section with a sub-section per linked resource.

    e.  **Save the Output:**
        *   Use `run_shell_command` to save the synthesized Markdown to the target file. Store the content in a variable and write it using `printf`. For example: `CONTENT="..."; printf '%s' "$CONTENT" > path/to/description.md`.

4.  **Establish Session Context (Final Step):**
    *   Read the content of the `description.md` file you just created using `run_shell_command` with `cat`.
    *   Read the content of `AGENTS.md` from the project root (fall back to `GEMINI.md` if not present).
    *   Format and display using the following Markdown structure EXACTLY, printing it to standard output:

        ```markdown
        ### ✨ Session Context Loaded for `{{feature_name}}`

        **Description:**
        > {{The synthesized description from the new description.md}}

        This context is now available for all subsequent commands.
        ```
    *   After outputting this block, the command is complete.
