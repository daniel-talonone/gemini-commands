---
name: tdd-skill
description: Guides the implementation of a feature using a strict Test-Driven Development (TDD) process. It uses a structured plan.yml file for requirements and state, and the yq tool for state updates.
---

# TDD (Test-Driven Development) Workflow

Your primary directive is to follow a strict Red-Green-Refactor cycle. The `plan.yml` file is the single source of truth, and all state modifications **must** be done using the `yq` command-line tool.

## 1. Pre-computation: Identify Project Context

1.  **Identify Test Command:** Find the project's testing script.
2.  **Identify Feature Directory:** Find the active feature directory and the `plan.yml` file within it.

## 2. TDD Execution Loop

### STEP 1: Select the Task

1.  Use `yq` to find the first task in `plan.yml` with `status == "todo"` or `status == "in-progress"`.
2.  Confirm the selected task with the user.

### STEP 2: Deconstruct or Resume Task

1.  Use `yq` to check if the active task object has a `sub_tasks` field.
2.  If not, deconstruct the task into a sub-task list, get user approval, and use a `yq` command to add the `sub_tasks` list and update the parent task's status to `in-progress`.

### STEP 3: RED (Write a Failing Test)

1.  Use `yq` to find the first sub-task with `status == "todo"`.
2.  If the task is ambiguous, read `description.md` for context.
3.  Write and verify a failing test.

### STEP 4: GREEN (Make the Test Pass)

1.  Write the minimum code to make the test pass.
2.  Verify it passes.

### STEP 5: REFACTOR (Clean Up the Code)

1.  Refactor the code and tests, ensuring tests still pass.

### STEP 6: CHECKPOINT AND REPEAT

1.  **Update `plan.yml` using `yq`:**
    *   Construct the correct `yq -i '...' plan.yml` command to find the sub-task you just completed by its `id` and set its `status` to `'done'`. Execute it with `run_shell_command`.
    *   After the sub-task is marked, read the file again with `yq` to check if all other sub-tasks for the parent task are also `'done'`.
    *   If they are, construct and execute a second `yq` command to set the parent task's `status` to `'done'`.
2.  **Check for more work:**
    *   If there are more incomplete sub-tasks, return to **STEP 3 (RED).**
    *   If all sub-tasks are complete, return to **STEP 1** to select the next high-level task.
3.  If all high-level tasks are complete, inform the user.

## Flexibility and Overrides

The user is in control. If they issue a direct command, it overrides this workflow.
