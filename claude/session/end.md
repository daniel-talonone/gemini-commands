---
description: Ends the work session, saving progress to the feature directory and project-wide knowledge to AGENTS.md.
---

You are an assistant that helps finalize a work session. Your primary task is to persist the *rationale* behind all important knowledge gained during this session.

**Actions:**

1.  **Identify Context:**
    *   Identify the active feature directory path from our conversation.
    *   Perform a final, comprehensive review of our entire session to determine which tasks were completed and which questions were answered.

2.  **Delegate Final State Update:**
    *   Delegate the final file modifications to a sub-agent (Agent tool, subagent_type: "general-purpose").
    *   The request must contain:
        *   The path to the feature directory.
        *   A list of all remaining task IDs to be marked as 'done' (or their final status).
        *   A list of all remaining question objects that were answered, each containing an `id` and the `answer`.
    *   The instruction payload for the sub-agent should be a prompt like this:
        """
        You are a YAML file specialist performing a final update. Your task is to update session state files using `yq`.
        1. For each task ID provided, update its status in `plan.yml` using the Bash tool.
           Example: `yq -i '(.[] | select(.id == "task-id")).status = "done"' plan.yml`
        2. For each question object provided, update its status to 'resolved' and set its `answer` in `questions.yml`.
           Example: `yq -i '(.[] | select(.id == "question-id")).status = "resolved" | (.[] | select(.id == "question-id")).answer = "The answer text."' questions.yml`
        3. Confirm when you have successfully updated all files.
        """

3.  **Generate and Save Final Log Summary:**
    *   Generate a concise but comprehensive "Session Summary" Markdown entry based on all the work completed.
    *   Call the `append_to_log.sh` script using the Bash tool to append it to `log.md`.
    *   Example: `$AI_SESSION_HOME/scripts/append_to_log.sh "path/to/your/log.md" "Your generated summary text."`

4.  **Update Project Knowledge (`AGENTS.md`):**
    *   Distill any **new, permanent, project-wide knowledge** and its rationale from the session.
    *   Find the `### ✨ Session Context Loaded for...` block in the conversation history and get the **Project Conventions** content from it.
    *   Integrate the new knowledge into this content and use the Edit tool to update the `AGENTS.md` file in the project root (fall back to `GEMINI.md` if `AGENTS.md` does not exist).

5.  **Confirm:** Announce that the session is complete and that both the feature directory and the project knowledge base have been updated.
