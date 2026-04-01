---
description: Ends the work session, saving progress to the feature directory and project-wide knowledge to AGENTS.md.
---

You are an assistant that helps finalize a work session. Your primary task is to persist the *rationale* behind all important knowledge gained during this session.

**Actions:**

1.  **Identify Context:**
    *   Identify the active feature directory path from our conversation.
    *   Perform a final, comprehensive review of our entire session to determine which tasks were completed and which questions were answered.

2.  **Update State Files:**
    *   For each remaining task ID, run the `yq` command using the Bash tool to update its final status in `plan.yml`.
        Example: `yq -i '(.[] | select(.id == "task-id")).status = "done"' path/to/plan.yml`
    *   For each answered question, run the `yq` command to update its status to 'resolved' and set its `answer` in `questions.yml`.
        Example: `yq -i '(.[] | select(.id == "question-id")).status = "resolved" | (.[] | select(.id == "question-id")).answer = "The answer text."' path/to/questions.yml`

3.  **Generate and Save Final Log Summary:**
    *   Generate a concise but comprehensive "Session Summary" Markdown entry based on all the work completed.
    *   Call the `append_to_log.sh` script using the Bash tool to append it to `log.md`.
        Example: `$AI_SESSION_HOME/scripts/append_to_log.sh "path/to/your/log.md" "Your generated summary text."`

4.  **Update Project Knowledge (`AGENTS.md`):**
    *   Distill any **new, permanent, project-wide knowledge** and its rationale from the session.
    *   Find the `### ✨ Session Context Loaded for...` block in the conversation history and get the **Project Conventions** content from it.
    *   Integrate the new knowledge into this content and use the Edit tool to update the `AGENTS.md` file in the project root (fall back to `GEMINI.md` if `AGENTS.md` does not exist).

5.  **Confirm:** Announce that the session is complete and that both the feature directory and the project knowledge base have been updated.
