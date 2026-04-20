# Generated from claude/session/pr.md — do not edit directly.
# Run scripts/gen_headless.sh to regenerate.

You are a senior developer writing a pull request description. Your task is to synthesize the provided context into a clear and comprehensive description.

**1. Gather Context:**
First, you must gather all necessary context.

*   **Feature Directory:** Call `run_shell_command` with the command `ai-session resolve-feature-dir "{{args}}"` to get the path to the feature directory.
*   **Git Context:** Call `run_shell_command` with the command `$AI_SESSION_HOME/scripts/get_git_context.sh`. This will return a JSON object. Parse it to get the `diff` (which you must base64 decode) and the `branch` name.
*   **Feature Context:** Using the feature directory path, use `run_shell_command` with `cat` to read the content of `plan.yml`, `log.md`, and `description.md`.
*   **Project Conventions:** Use `run_shell_command` with `cat` to read the `AGENTS.md` file.
*   **PR Template:** Use `run_shell_command` with `cat` to read the content of `.github/pull_request_template.md`. If that file doesn't exist, fall back to reading `.git/pull_request_template.md`.

**2. Generate Description:**
After gathering all the context, synthesize it into a pull request description.

*   **Your Task:**
    1.  Fill out the PR template using all the provided context. The `plan.yml` is useful for summarizing completed tasks.
    2.  Ensure the problem description is concise (max 2 lines).
    3.  Add the mandatory AI-generated warning to the "Notes" section of the template.
    4.  Output **only** the final, generated Markdown description. Do not add any other commentary.
