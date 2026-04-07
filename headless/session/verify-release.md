# Generated from claude/session/verify-release.md — do not edit directly.
# Run scripts/gen_headless.sh to regenerate.

Your task is to run a verification script and provide a semantic analysis of its JSON output to ensure the integrity of a release branch.

The arguments for the verification are: `{{args}}`

**Steps:**

1.  **Execute Verification Script:**
    *   Use the `run_shell_command` tool to execute the `verify-release.sh` script, passing the arguments to it.
    *   Command: `$AI_SESSION_HOME/scripts/verify-release.sh {{args}}`
    *   This script compares original commits to the cherry-picked commits and produces a JSON report.

2.  **Analyze the JSON Output:**
    *   The script's entire output will be a single JSON object. Parse this JSON.
    *   Check the `status` field:
    *   **If the status is "VERIFICATION_SUCCESSFUL"**: Output "VERIFICATION_SUCCESSFUL" and stop.
    *   **If the status is "VERIFICATION_FAILED"**: Parse the other fields and generate a clear, analyzed report covering:
        *   **Extra Commits**: List any commit hashes from `extra_commits` if non-empty.
        *   **Missing Commits**: List any commit hashes from `missing_commits` if non-empty.
        *   **Changed Commits**: For each object in `changed_commits` (which has `original_commit`, `release_commit`, and `diff` fields), perform a semantic analysis of the `diff` and provide a concise summary answering:
            1.  **What was the nature of the change?**
            2.  **Why did it likely change?**
            3.  **What is the risk assessment?** (Rate **Low**, **Medium**, or **High** with a brief justification.)

3.  **Present Final Report:**
    *   If verification failed, combine all findings into a single, well-structured Markdown report.
    *   Start the report with "VERIFICATION FAILED".
    *   The report must have the following sections:
        *   `### 📝 Changed Commits Analysis`
        *   `### ⚠️ Extra Commits`
        *   `### ❓ Missing Commits`
    *   Do not output the raw `diff` strings from the JSON, only your analysis.
    *   Output the final report to stdout.
