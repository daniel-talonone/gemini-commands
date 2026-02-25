---
name: tdd-skill
description: Guides the implementation of a feature using a strict Test-Driven Development (TDD) process. It uses an active feature document for requirements and state.
---

# TDD (Test-Driven Development) Workflow

Your primary directive is to follow a strict Red-Green-Refactor cycle for software development. You must adhere to the following procedure step-by-step. The Feature Document is the single source of truth for all state.

## 1. Pre-computation: Identify Project Context

Before starting the loop, you must identify the necessary context for the project.

1.  **Identify Test Command:** Analyze the project to find the testing script (e.g., in `package.json`). If you cannot find it, you **must** ask the user.
2.  **Identify Feature Document:** Identify the active feature document (e.g., `.vscode/sc-*.md`) from the conversation history. If unsure, **ask the user**.

## 2. TDD Execution Loop

### STEP 1: Select the Task

1.  Read the feature document.
2.  In the "## Next Steps" section, find the first high-level task that is not marked `[x]`.
3.  **Confirm with the user:** "I am now starting work on the following task: '[task description]'. Is this correct?" Do not proceed without confirmation.

### STEP 2: Deconstruct or Resume Task

1.  Examine the selected task in the feature document.
2.  **If a sub-task checklist does NOT exist under the main task:**
    *   Break the high-level task down into a checklist of smaller, concrete, testable sub-tasks (e.g., `- [ ] Test for...`).
    *   Present this sub-task checklist to the user for approval.
    *   Once approved, **use the `replace` tool to insert this checklist directly below the main task in the feature document.** Indent the sub-tasks.
3.  **If a sub-task checklist already exists,** inform the user you are resuming work from the existing plan.

### STEP 3: RED (Write a Failing Test)

1.  **Read the feature document again.** Identify the **first incomplete sub-task** (`- [ ] ...`) from the checklist under the active high-level task.
2.  Based on this sub-task, write the simplest possible test that will fail.
3.  **Present the test code and file path to the user for approval.**
4.  Once approved, use `write_file` or `replace` to add the test.
5.  Execute the identified test command using `run_shell_command`.
6.  **Confirm that the test fails for the expected reason.** If it passes or fails unexpectedly, you must debug the test itself before proceeding.

### STEP 4: GREEN (Make the Test Pass)

1.  Write the **absolute minimum amount of code** to make the test pass.
2.  **Present the implementation code to the user for approval.**
3.  Once approved, write the code to the codebase.
4.  Execute the test command again and **confirm that the test now passes.** Debug until it does.

### STEP 5: REFACTOR (Clean Up the Code)

1.  Critically review the application and test code for this sub-task.
2.  Propose specific refactoring changes to the user.
3.  If approved, apply the changes and run tests a final time to ensure they all still pass.

### STEP 6: CHECKPOINT AND REPEAT

1.  **Update the Feature Document:** Use the `replace` tool to find the line for the sub-task you just completed and change it from `- [ ]` to `- [x]`.
2.  **Check for more work:**
    *   Read the feature document again. If there are more incomplete sub-tasks for the current high-level task, **return to STEP 3 (RED).**
    *   If all sub-tasks are complete, mark the main high-level task as `[x]` in the feature document. Then, **return to STEP 1** to select the next high-level task.
3.  If all high-level tasks in "Next Steps" are complete, inform the user.

## Flexibility and Overrides

The user is in control. If they issue a direct command, it overrides this workflow.
