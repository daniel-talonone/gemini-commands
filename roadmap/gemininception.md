# Initiative: The Focused Task Pattern

## The Goal (The "Why")

The core architectural challenge in a long-running, stateful agent session is managing context. As more information (files, chat history, tool outputs) is loaded, the context grows, leading to two primary problems:

1.  **Token Inefficiency:** Every conversational turn requires sending the entire context to the LLM. A large context is expensive and slow.
2.  **Context Pollution:** The LLM's attention is diluted. When asked to perform a specific task (e.g., summarize an article), the unrelated context of the main session can degrade the quality and focus of its output.

The **Focused Task Pattern** is the solution to this problem. It's an evolution of the project's core philosophy: using the right tool for the job. In this case, the "tool" is a temporary, specialized sub-agent with a precisely defined, minimal context.

## The Solution: The Subagent Pattern

The CLI's built-in **Subagent** feature provides a structured, robust, and maintainable solution for delegating complex tasks. For more details, see the [official documentation on subagents](https://geminicli.com/docs/core/subagents/).

### What are Subagents?
Subagents are specialized agents that the main agent can call like any other tool. They have their own isolated context, history, and toolset, achieving the same goal of context isolation without manual shell scripting.

### The `generalist` Subagent
The `generalist` is a built-in subagent that is perfectly suited for implementing the Focused Task Pattern. It is a general-purpose agent that has access to all tools and can be given complex, multi-step instructions.

### The Pattern Explained
The workflow is simple:

1.  A command's `.toml` file contains a prompt that instructs the main agent to delegate a task.
2.  The main agent calls the `generalist` subagent tool, passing it the detailed instructions for the task.
3.  The `generalist` subagent executes the entire task in its own isolated session, using its own tools (like `run_shell_command`) as needed.
4.  The final result is returned to the main agent.

---

## Potential Use Cases

This pattern is highly versatile. Here are several concrete applications:

### Use Case 1: Processing Large or Volatile Data
**Problem:** A task requires processing a large or frequently changing piece of data, like a `git diff`. Adding this to the main context would be expensive and quickly become stale.
**Solution:** A dedicated command (e.g., `/session:get-familiar`) can use the `generalist` subagent to get the diff, summarize it, and present the small, cheap summary to the user, without ever polluting the main context.

### Use Case 2: Unbiased Analysis
**Problem:** The agent that wrote the code is asked to review it. Its knowledge of the implementation history can create blind spots, leading to a biased, less critical review.
**Solution:** The `/session/review` command can use the `generalist` subagent to create a "fresh eyes" reviewer. The main agent gathers the *requirements* (`description.md`, `questions.yml`) and the final *code* (`git diff`), but deliberately **excludes** the implementation history (the main chat log). It passes this objective context to the sub-agent, forcing an unbiased review of the code against its requirements.

### Use Case 3: Content Generation from Specific Inputs
**Problem:** A command needs to generate a structured piece of text (like a PR description or a research log) based on a few specific inputs.
**Solution:** Instead of using the full session context, commands like `/session:pr` or `/session:log-research` can use the `generalist` subagent to gather only the relevant inputs (e.g., plan file, git context, or a single web article), and delegate the generation task to a focused sub-session. This is faster, cheaper, and produces more predictable results.

---

## Example: `/session:get-familiar`
The implementation for this command becomes a simple instruction to the main agent.

The `prompt` in `session/get-familiar.toml`:
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
