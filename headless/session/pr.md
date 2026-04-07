# Generated from claude/session/pr.md — do not edit directly.
# Run scripts/gen_headless.sh to regenerate.

You are an orchestrator for creating pull requests. You will gather all necessary context, generate the description, and then handle the GitHub interaction and state updates.

**Process:**

1.  **Gather All Context:**
    *   **Feature Directory:** Resolve the feature directory path from the input argument.
        ```bash
        FEATURE_DIR="$($AI_SESSION_HOME/scripts/resolve_feature_dir.sh "{{args}}")"
        ```
    *   **Git Context:** Call `$AI_SESSION_HOME/scripts/get_git_context.sh` to get the diff and the current branch name. Decode the diff from base64.
        ```bash
        # Note: This script outputs branch and base64-encoded diff on separate lines.
        GIT_CONTEXT_OUTPUT=$($AI_SESSION_HOME/scripts/get_git_context.sh)
        BRANCH_NAME=$(echo "$GIT_CONTEXT_OUTPUT" | head -n 1)
        ENCODED_DIFF=$(echo "$GIT_CONTEXT_OUTPUT" | tail -n +2)
        DECODED_DIFF=$(echo "$ENCODED_DIFF" | base64 -d)
        ```
    *   **Feature Context:** Read the content of `plan.yml`, `log.md`, and `description.md` from the resolved feature directory.
        ```bash
        PLAN_YML_CONTENT=$(cat "$FEATURE_DIR/plan.yml")
        LOG_MD_CONTENT=$(cat "$FEATURE_DIR/log.md")
        DESCRIPTION_MD_CONTENT=$(cat "$FEATURE_DIR/description.md")
        ```
    *   **Project Conventions:** Read the `AGENTS.md` file.
        ```bash
        AGENTS_MD_CONTENT=$(cat AGENTS.md)
        ```
    *   **PR Template:** Read the content of `.git/pull_request_template.md`. If it doesn't exist, proceed with an empty string.
        ```bash
        PR_TEMPLATE_CONTENT=""
        if [ -f .git/pull_request_template.md ]; then
            PR_TEMPLATE_CONTENT=$(cat .git/pull_request_template.md)
        fi
        ```

2.  **Generate Description:**
    *   Synthesize the gathered context into a comprehensive pull request description.

    You are a senior developer writing a pull request description. Your task is to synthesize the provided context into a clear and comprehensive description.

    **Provided Context:**
    *   **PR Template Content:** `{{pr_template_content}}`
    *   **Feature Description:** `{{content of description.md from session context}}`
    *   **Implementation Plan:** `{{content of plan.yml}}`
    *   **Development Log:** `{{content of log.md}}`
    *   **Code Diff:** `{{decoded diff}}`
    *   **Project Conventions:** `{{content of AGENTS.md from session context}}`

    **Task:**
    1.  Fill out the PR template using all the provided context. The `plan.yml` is useful for summarizing completed tasks.
    2.  Ensure the problem description is concise (max 2 lines).
    3.  Add the mandatory AI-generated warning to the "Notes" section of the template.
    4.  Output **only** the final, generated Markdown description. Do not add any other commentary.

3.  **Find, Create or Update Pull Request:**
    *   Write the generated description to a temporary file.
    *   Check if a pull request for the current branch already exists.
    *   If a PR exists, update its body with the new description.
    *   If no PR exists, create a new pull request using the branch name as the title.
        ```bash
        # Write the generated description to a temp file
        PR_BODY_FILE=$(mktemp)
        printf '%s' "$GENERATED_PR_DESCRIPTION" > "$PR_BODY_FILE"

        # Check for existing PR
        EXISTING_PR_URL=$(gh pr list --head "$BRANCH_NAME" --json url -q '.[0].url')

        if [ -n "$EXISTING_PR_URL" ]; then
            # Update existing PR
            gh pr edit "$EXISTING_PR_URL" --body-file "$PR_BODY_FILE"
            PR_URL="$EXISTING_PR_URL"
        else
            # Create new PR
            PR_URL=$(gh pr create --title "$BRANCH_NAME" --body-file "$PR_BODY_FILE")
        fi

        # Clean up temp file
        rm "$PR_BODY_FILE"
        ```

4.  **Save PR Link:**
    *   Append the URL of the created or updated pull request to `description.md` in the feature directory.
        ```bash
        printf '\n%s\n' "$PR_URL" >> "$FEATURE_DIR/description.md"
        ```
