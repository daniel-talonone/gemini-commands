# Refactoring Plan: Migrating to Helper Scripts

## Goal

To improve the reliability and determinism of the session commands by migrating procedural, deterministic logic out of LLM prompts and into dedicated, executable helper scripts.

## Architectural Rationale (The "Why")

The core philosophy of this project is to use robust, deterministic tools for state updates and procedural tasks, reserving the LLM for creative and synthesis-based tasks. We identified that asking the LLM to perform multi-step procedures (e.g., creating a directory, then creating multiple files inside it) is brittle. The LLM is non-deterministic and can fail, misinterpret, or forget steps, leading to inconsistent command execution.

This refactoring effort is a direct continuation of the project's architectural evolution:
1.  **From Single File** -> **To Feature Directory** (for reliable file operations)
2.  **From Markdown Lists** -> **To Structured YAML** (for reliable data parsing)
3.  **From In-Memory Parsing** -> **To `yq` CLI** (for atomic state updates)
4.  **From LLM Procedures** -> **To Helper Scripts** (for reliable procedural actions)

The solution is to isolate this procedural logic into simple, robust shell scripts. The LLM's role shifts from *performing* the procedure to *calling* the script that performs the procedure.

## The Script Execution Pattern (The "How")

A key challenge was executing these scripts, which are stored centrally, from any user directory. We established and validated the following pattern:

-   **Problem:** Scripts are located in `~/.gemini/commands/scripts/`, but session commands are executed from arbitrary project directories.
-   **Solution:** The LLM prompt will construct an absolute path to the script at runtime by leveraging the known conventional path for global user commands: `$HOME/.gemini/commands/`.

The command inside a prompt looks like this:
```bash
#!/bin/bash
COMMANDS_DIR="$HOME/.gemini/commands"
"$COMMANDS_DIR/scripts/your_script_name.sh" "arg1" "arg2"
```
This pattern is self-contained, requires no user setup (unlike a `setup.sh` script), and is resilient.

## Completed Work

This refactoring pattern has been successfully applied to several commands:

-   **/session:new**: Uses `scripts/create_feature_dir.sh` to scaffold the feature directory from a Shortcut story ID. The LLM's only file-writing task is to synthesize and write `description.md`.
-   **/session:checkpoint**, **/session:end**, and **/session:log-research**: These commands now use the `scripts/append_to_log.sh` script to add timestamped entries to the `log.md` file. This ensures consistent formatting and avoids race conditions or file corruption from a "Read-Append-Write" pattern being handled by the LLM.
-   **/session:migration**: Was refactored to use `scripts/migrate_feature_file.sh`. This replaced a complex, brittle, multi-step prompt with a single, robust script that handles the entire file and directory migration atomically, significantly improving reliability.
-   **/session:start** & **/session:summary**: These commands were refactored to use `scripts/load_context_files.sh`. This replaced ~6 individual `read_file` tool calls with a single script execution, improving the performance and reducing the token overhead for starting a session or generating a summary.

### High Priority

-   **/session:define**:
    -   **Current Logic:** The prompt instructs the LLM to generate a directory name, create a directory, and then create multiple placeholder files (`plan.yml`, `log.md`, etc.).
    -   **Proposed Script:** Reuse `scripts/create_feature_dir.sh`.
    -   **Implementation:** The LLM's role should be focused on the conversation to define the feature. Once approved, it should generate the directory name and call the `create_feature_dir.sh` script. The LLM's final step would be a single `write_file` call to create `description.md` with the synthesized content, mirroring the robust pattern used in `/session:new`.

### Medium Priority

-   **/session:address-feedback**:
    -   **Current Logic:** The prompt uses a "Read-Append-Write" pattern to log the rationale for a fix to `log.md`.
    -   **Proposed Script:** Reuse `scripts/append_to_log.sh`.
    -   **Implementation:** The LLM should generate the summary of the feedback and the rationale, then call the script to append it to the log, just as `/session:checkpoint` does.

### Low Priority

-   **/session:pr**, **/session:pr_from_branch**, **/session:review**, & **/session:review_from_branch**:
    -   **Current Logic:** These prompts all involve running multiple `git` commands to get context (current branch name, main branch name, diff against main).
    -   **Proposed Script:** `scripts/get_git_context.sh`.
    -   **Implementation:** A single script could reliably gather all necessary git context and return it as a structured string (e.g., JSON), which the LLM can then use for generating the PR description or performing the code review. This would consolidate the logic and reduce the number of shell commands the LLM needs to invoke.
    -   **Future Consideration:** The output from `get_git_context.sh` is currently piped directly to the next command. For debugging or handling extremely large diffs, it might be beneficial to store this JSON output in a file within the feature directory (e.g., `git_context.json`). This would create a persistent artifact of the context used, but introduces the risk of the data becoming stale if not regenerated before each use. This is a trade-off to be evaluated later.


## Future Refactoring Opportunities (TODO)

## Architectural Pattern: The "Focused Task" Sub-Session (or "Gemini Inception")

Beyond simple script migration, a more advanced architectural pattern has been identified for improving the efficiency and reliability of commands. This pattern moves away from a single, monolithic session context and towards a main **orchestrator** session that can spawn temporary, isolated **sub-sessions** for specific tasks.

### Core Principle

The main session should not perform every task itself. For tasks that do not require the full conversation history or session context (like summarizing a file or generating a description from specific inputs), the main session's role is to:

1.  **Prepare a minimal, focused context** required for that single task.
2.  **Delegate the task** to an isolated sub-session by piping this minimal context into a `gemini query "..."` command.
3.  **Use the result** from the sub-session, which then terminates, freeing its memory and context.

This prevents the main session's context from being polluted with large, one-off data blobs (like a `git diff`) and significantly reduces token consumption by ensuring only the necessary information is sent for each step.

### Example: A Token-Efficient `/session:pr` command

Instead of using the main session, the command's prompt would be a script that orchestrates a sub-session:

```bash
#!/bin/bash

# 1. The orchestrator gathers the PRECISE context needed.
DIFF_JSON=$(sh "$HOME/.gemini/commands/scripts/get_git_context.sh")
DESCRIPTION=$(cat .vscode/{{args}}/description.md)
PLAN=$(cat .vscode/{{args}}/plan.yml)

# 2. It prepares a focused, one-shot prompt for the sub-session.
FOCUSED_PROMPT="""
Write a standard pull request body based on the following plan, feature description, and code diff.
PLAN:
$PLAN

DESCRIPTION:
$DESCRIPTION

DIFF:
$(echo "$DIFF_JSON" | jq -r '.diff' | base64 --decode)
"""

# 3. It spawns the "Inception" session to do one job.
PR_BODY=$(echo "$FOCUSED_PROMPT" | gemini query)

# 4. The main session uses the result to create the PR.
# The sub-session and its large context (diff, etc.) are gone.
gh pr create --title "feat({{args}}): Implement feature" --body "$PR_BODY"
```

### Candidates for this Pattern

-   **/session:review_changes (New)**: A new command to summarize the current git diff without adding it to the main context.
-   **/session:log-research**: The web fetching and summarizing can be done in an isolated sub-session that is only given the article's content.
-   **/session:pr** & **/session:pr_from_branch**: Can use this pattern to generate the PR body with a precisely controlled context, as shown in the example.
