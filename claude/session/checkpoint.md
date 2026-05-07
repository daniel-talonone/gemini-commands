---
description: Saves a checkpoint of the work done by updating the state files.
---

You are an assistant that helps log work-in-progress and update the session state.

**Actions:**

1.  **Identify Context:**
    *   Identify the active feature directory path from our conversation.
    *   Review our conversation since the last checkpoint to determine which tasks were completed and which questions were answered.

2.  **Update State Files:**
    *   For each completed task ID, run the `yq` command using the Bash tool to update its status to 'done' in `plan.yml`.
        Task update: `yq -i '(.[] | .tasks[] | select(.id == "task-id")).status = "done"' path/to/plan.yml`
        Slice update: `yq -i '(.[] | select(.id == "slice-id")).status = "done"' path/to/plan.yml`
    *   For each answered question, run the `yq` command to update its status to 'resolved' and set its `answer` in `questions.yml`.
        Example: `yq -i '(.[] | select(.id == "question-id")).status = "resolved" | (.[] | select(.id == "question-id")).answer = "The answer text."' path/to/questions.yml`

3.  **Update Log (`log.md`):**
    *   Generate a concise Markdown summary of the work done.
    *   Resolve the feature directory path, then append the log:
        ```bash
        FEATURE_DIR=$(ai-session resolve-feature-dir "<feature-id>")
        ai-session append-log "$FEATURE_DIR" "Your generated summary text."
        ```

4.  **Confirm:** State that the checkpoint has been saved.
