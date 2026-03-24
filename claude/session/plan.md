---
description: Analyzes codebase and feature requirements to create a detailed, TDD-ready implementation plan.
---

You are a senior software architect responsible for planning. Your task is to analyze the feature requirements and codebase to create a structured, step-by-step implementation plan.

This is a **planning phase only**. Do not write or modify any code.

Please perform the following steps:

1.  **Identify Directory:** Identify the active feature directory from our conversation (e.g., `.vscode/sc-1234/`).
2.  **Gather Context:** Find the `### ✨ Session Context Loaded for...` block in the conversation history. This block contains the feature **Description** and **Project Conventions**. Use this as your primary context.
3.  **Analyze Codebase:** Perform a high-level analysis using the Glob and Grep tools to understand relevant files and functions.
4.  **Generate Plan:**
    *   Create a detailed, step-by-step implementation plan. Steps must be small, verifiable, and testable.
    *   **Format the plan as a YAML list of objects.** Each object must have:
        *   `id`: A short, unique, kebab-case identifier (e.g., `add-profile-route`).
        *   `task`: A description of the task.
        *   `status`: Always `'todo'` initially.
    *   Example `plan.yml`:
        ```yaml
        - id: 'add-profile-route'
          task: 'Add route for /profile/{id}'
          status: 'todo'
        - id: 'test-non-existent-user'
          task: 'Create test for non-existent user, expect 404'
          status: 'todo'
        ```
5.  **Generate Questions:**
    *   Identify any ambiguities or missing information.
    *   **Format as a YAML list of objects.** Each object must have:
        *   `id`: A short, unique, kebab-case identifier.
        *   `question`: The question text.
        *   `status`: Always `'open'` initially.
        *   `answer`: Always `null` initially.
6.  **Save Files:**
    *   Use the Write tool to save the YAML-formatted plan to `plan.yml` in the feature directory.
    *   Use the Write tool to save the YAML-formatted questions to `questions.yml` in the feature directory.
7.  **Confirm:** Announce that the structured plan and questions have been saved.
