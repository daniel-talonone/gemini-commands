# Initiative: The Focused Task Pattern (or "Gemini Inception")

## The Goal (The "Why")

The core architectural challenge in a long-running, stateful agent session is managing context. As more information (files, chat history, tool outputs) is loaded, the context grows, leading to two primary problems:

1.  **Token Inefficiency:** Every conversational turn requires sending the entire context to the LLM. A large context is expensive and slow.
2.  **Context Pollution:** The LLM's attention is diluted. When asked to perform a specific task (e.g., summarize an article), the unrelated context of the main session can degrade the quality and focus of its output.

The **Focused Task Pattern** is the solution to this problem. It's an evolution of the project's core philosophy: using the right tool for the job. In this case, the "tool" is a temporary, specialized agent with a precisely defined, minimal context.

## The Pattern Explained (The "How")

> **Note:** This pattern describes a manual way to create isolated sessions. A more robust, modern approach is to use the built-in **`generalist` subagent**, as described in the "A More Robust Solution: The Subagent Pattern" section below.

Instead of having the main, stateful LLM perform every task, we use a `prompt` that is a `bash` script. This **Orchestrator Script** acts as a controller that can spawn temporary, isolated sub-sessions for specific AI tasks.

The pattern is a four-step process:

1.  **Orchestrate:** The main script gathers the *exact*, minimal data required for a specific task (e.g., reading a single file, getting the output of a git command).
2.  **Focus:** It prepares a new, single-purpose prompt that contains only the instructions and the data for the task at hand.
3.  **Delegate:** It pipes this focused prompt into a `gemini -p ""` command. This creates a new, non-interactive, temporary sub-session that is completely isolated from the main session's history and state.
4.  **Act:** The orchestrator script captures the result from the sub-session and can then use it for subsequent actions. The sub-session and its context are destroyed upon completion.

### Generic Example

This pseudo-code illustrates the pattern without tying it to a specific use case:

```bash
#!/bin/bash

# 1. ORCHESTRATE: Gather specific data needed for the task.
DATA_1=$(cat some_file.txt)
DATA_2=$(run_some_script.sh)

# 2. FOCUS: Prepare a precise, minimal prompt for the sub-session.
FOCUSED_PROMPT="""
Based on the following data, perform a specific generative task.
Data A: $DATA_1
Data B: $DATA_2
"""

# 3. DELEGATE: Execute the task in an isolated sub-session.
TASK_RESULT=$(echo "$FOCUSED_PROMPT" | gemini -p "")

# 4. ACT: Use the result in the main session.
echo "Sub-session complete. Result: $TASK_RESULT"
```

---

## Potential Use Cases

This pattern is highly versatile. Here are several concrete applications:

### Use Case 1: Processing Large or Volatile Data
**Problem:** A task requires processing a large or frequently changing piece of data, like a `git diff`. Adding this to the main context would be expensive and quickly become stale.
**Solution:** A dedicated command (e.g., `/session:get_familiar`) can use the pattern to get the diff, pipe it to a sub-session for summarization, and present the small, cheap summary to the user, without ever polluting the main context.

### Use Case 2: Unbiased Analysis
**Problem:** The agent that wrote the code is asked to review it. Its knowledge of the implementation history can create blind spots, leading to a biased, less critical review.
**Solution:** The `/session/review` command can use the pattern to create a "fresh eyes" reviewer. The orchestrator script gathers the *requirements* (`description.md`, `questions.yml`) and the final *code* (`git diff`), but deliberately **excludes** the implementation history (the main chat log). It pipes this objective context to a sub-session, forcing an unbiased review of the code against its requirements.

### Use Case 3: Content Generation from Specific Inputs
**Problem:** A command needs to generate a structured piece of text (like a PR description or a research log) based on a few specific inputs.
**Solution:** Instead of using the full session context, commands like `/session:pr` or `/session:log-research` can use the pattern to gather only the relevant inputs (e.g., plan file, git context, or a single web article), and delegate the generation task to a focused sub-session. This is faster, cheaper, and produces more predictable results.

---

## A More Robust Solution: The Subagent Pattern

While the "Focused Task Pattern" is effective, the CLI's built-in **Subagent** feature provides a more structured, robust, and maintainable solution for delegating complex tasks. For more details, see the [official documentation on subagents](https://geminicli.com/docs/core/subagents/).

### What are Subagents?
Subagents are specialized agents that the main agent can call like any other tool. They have their own isolated context, history, and toolset, achieving the same goal of context isolation without manual shell scripting.

### The `generalist` Subagent
The `generalist` is a built-in subagent that is perfectly suited for the use cases previously handled by the "Gemini Inception" pattern. It is a general-purpose agent that has access to all tools and can be given complex, multi-step instructions.

### The Updated Pattern
The workflow becomes much simpler:

1.  A command's `.toml` file contains a prompt that instructs the main agent to delegate a task.
2.  The main agent calls the `generalist` subagent tool, passing it the detailed instructions for the task.
3.  The `generalist` subagent executes the entire task in its own isolated session, using its own tools (like `run_shell_command`) as needed.
4.  The final result is returned to the main agent.

This approach is preferred as it is cleaner and uses a built-in feature, removing the need for manual process management with `gemini -p`.

### Example: `/session:get_familiar`
The implementation for this command becomes a simple instruction to the main agent.

The `prompt` in `session/get_familiar.toml`:
```toml
description = "Gets familiar with the current code changes by having a subagent generate a summary."
ephemeral = true
prompt = """
Use the `generalist` subagent to perform a review of the current code changes.

Pass the following prompt to the generalist:
"You are an expert code reviewer. Your task is to provide a concise summary of the code changes on the current branch.

Follow these steps:
1.  Execute the shell script using its full, absolute path: `$HOME/.gemini/commands/scripts/get_git_context.sh`.
2.  The JSON output contains a `diff` field which is base64 encoded. You must decode this diff content before analyzing it.
3.  If the decoded diff is empty, inform the user that there are no changes to review and stop.
4.  If there are changes, analyze the decoded diff and provide a high-level, concise summary."
"""
```
