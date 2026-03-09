# Refactoring Plan: Migrating to Helper Scripts

## Goal

To improve the reliability and determinism of the session commands by migrating procedural, deterministic logic out of LLM prompts and into dedicated, executable helper scripts.

## Architectural Rationale (The "Why")

The core philosophy of this project is to use robust, deterministic tools for state updates and procedural tasks, reserving the LLM for creative and synthesis-based tasks. We identified that asking the LLM to perform multi-step procedures (e.g., creating a directory, then creating multiple files inside it) is brittle. The LLM is non-deterministic and can fail, misinterpret, or forget steps, leading to inconsistent command execution.

This refactoring effort is a direct continuation of the project's architectural evolution:
1.  **From Single File** -> **To Feature Directory** (for reliable file operations)
2.  **From Markdown Lists** -> **To Structured YAML** (for reliable data parsing)
3.  **From In-Memory Parsing** -> **To `yq` CLI** (for atomic state updates)
4.  **From LLM Procedures** -> **To Helper Scripts** (for reliable procedural actions)

The solution is to isolate this procedural logic into simple, robust shell scripts. The LLM's role shifts from *performing* the procedure to *calling* the script that performs the procedure.

## The Script Execution Pattern (The "How")

A key challenge was executing these scripts, which are stored centrally, from any user directory. We established and validated the following pattern:

-   **Problem:** Scripts are located in `~/.gemini/commands/scripts/`, but session commands are executed from arbitrary project directories.
-   **Solution:** The LLM prompt will construct an absolute path to the script at runtime by leveraging the known conventional path for global user commands: `$HOME/.gemini/commands/`.

The command inside a prompt looks like this:
```bash
#!/bin/bash
COMMANDS_DIR="$HOME/.gemini/commands"
"$COMMANDS_DIR/scripts/your_script_name.sh" "arg1" "arg2"
```
This pattern is self-contained, requires no user setup (unlike a `setup.sh` script), and is resilient.

## Completed Work

We have successfully refactored the `/session:new` command as a proof-of-concept for this pattern.
-   `scripts/create_feature_dir.sh` was created.
    -   It accepts two arguments: `<base-dir>` and `<story-id>`.
    -   It creates the specified directory and populates it with the default placeholder files and content (`plan.yml`, `log.md`, etc.).
-   `session/new.toml` was updated to call this script. The LLM's only remaining file-writing task for this command is to synthesize and write `description.md`.

## Next Steps & Future Work (TODO)

The next user or LLM agent should continue this refactoring effort.

**Your task is to:**

1.  **Analyze** the remaining `.toml` files in the `session/` directory.
2.  **Identify** commands whose prompts contain procedural logic (especially file system operations like creating, moving, or modifying files in a sequence) that could be extracted into a script.
    -   A good candidate to start with might be `/session:migration`, which likely performs file migration operations.
3.  **For each identified command:**
    a. **Create** a new, robust shell script in the `scripts/` directory that encapsulates the procedural logic. Ensure the script is well-documented and handles parameters gracefully.
    b. **Update** the corresponding `.toml` file to remove the procedural instructions from the prompt.
    c. **Modify** the prompt to instead execute the new script using the established execution pattern detailed above.
