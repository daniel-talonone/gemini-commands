---
description: Logs a detailed, comprehensive summary of research findings to log.md using a sub-agent.
---

You are an orchestrator for logging research. Your goal is to review the session, create a summary of actions, and delegate the final report generation and file-writing to a sub-agent.

**Actions:**

1.  **Identify Directory:** Identify the active feature directory path from our conversation.

2.  **Summarize Actions:**
    *   Review the conversation since the last checkpoint or log entry.
    *   Create a concise, structured summary of all research and analysis activities. This is not the final report, but a bulleted list of facts for the next step. Include:
        *   The initial research question.
        *   Files read.
        *   Commands run.
        *   Web searches performed.
        *   Key insights or conclusions reached.

3.  **Delegate Report Generation:**
    *   Use the Agent tool (subagent_type: "general-purpose") to take your structured summary and generate the final, detailed log entry.
    *   Construct and pass the following detailed prompt to the sub-agent.

    ---
    **Sub-Agent Prompt:**

    You are a technical writer responsible for creating a detailed research log. You will be given a structured summary of actions and a file path.

    **Inputs:**
    *   **Action Summary:** `{{Structured summary created by the main agent}}`
    *   **Log File Path:** `{{path to the feature_directory}}/log.md`

    **Task:**
    1.  Based on the **Action Summary**, generate a **highly detailed, comprehensive report** in Markdown.
    2.  Do **not** summarize heavily; the goal is to expand on the provided points and preserve information for future reference. Elaborate on the "why" behind the actions.
    3.  Call the `$AI_SESSION_HOME/scripts/append_to_log.sh` script using the Bash tool to append your detailed report to the provided **Log File Path**.
    4.  Confirm that the log entry has been added. Do not output the report itself.
    ---

4.  **Confirm:** Announce that the research has been logged in detail.
