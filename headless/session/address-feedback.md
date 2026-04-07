# Generated from claude/session/address-feedback.md — do not edit directly.
# Run scripts/gen_headless.sh to regenerate.

Your task is to decompose an oversized user story into a set of maximally atomic sub-stories. Each sub-story must be scoped to a single concern, be implementable by an LLM without any human decision points, and be verifiable with a single build + test run.

The guiding principle is to bias toward smaller stories. An LLM can work more hours than a human; the cost of many tiny stories is near zero, while the cost of an ambiguous story is a failed autonomous run requiring human intervention.

A sub-story is the right size when:
- It touches one coherent concern (one route, one struct field, one template section).
- Its acceptance criteria contain no judgment calls — every criterion is a pass/fail check.
- An LLM can verify it with a single build and test run with no ambiguity.

This is a headless, non-interactive command. You will perform the analysis, create the sub-stories, and create all corresponding feature directories and files without any user interaction.

---

## 1. Load Context

The feature ID is provided in `{{args}}`.

First, resolve the feature directory path and read the necessary context files.
You MUST use `run_shell_command` for all file operations.

````bash
FEATURE_ID="{{args}}"
FEATURE_DIR=$(ai-session resolve-feature-dir "$FEATURE_ID")
DESCRIPTION=$(cat "$FEATURE_DIR/description.md")
REPO=$(yq e '.repo' "$FEATURE_DIR/status.yaml")
WORK_DIR=$(yq e '.work_dir' "$FEATURE_DIR/status.yaml")
````

Use the `run_shell_command` tool to execute these commands and store the output in variables for the subsequent steps.

---

## 2. Analyse the Story

Using the content of the `$DESCRIPTION` variable, analyze the user story. Identify:
- Which acceptance criteria are independent of each other.
- Which criteria introduce a new dependency (e.g., a schema change, a new library, a new route) that others will build on.
- The natural sequence of implementation: what must exist before the next thing can be added.

Based on this analysis, create a plan to split the story into a numbered list of sub-stories. For each sub-story, determine:
- A short **name** (kebab-case, will become the directory suffix, e.g., `empty-shell`).
- A one-sentence **summary** of what it delivers.
- Any dependencies on other sub-stories.

---

## 3. Create Feature Directories

For each sub-story you identified, you will now create its feature directory and `description.md` file.

Iterate through your list of sub-stories in order. For each one:

1.  Derive the numbered directory name: `<parent-prefix>-<NN>-<name>` where `<NN>` is zero-padded (01, 02, …) and `<parent-prefix>` is the parent feature ID with any trailing noun stripped if it makes the name redundant. Keep names concise.

2.  Compute the target path:
    ```bash
    PARENT_DIR=$(ai-session resolve-feature-dir "$FEATURE_ID")
    FEATURE_BASE=$(dirname "$PARENT_DIR")
    SUB_DIR="$FEATURE_BASE/<numbered-name>"
    ```

3.  Create the feature directory using the `ai-session` tool, inheriting `repo` and `work_dir` from the parent. The branch name should be the same as the numbered directory name.
    ```bash
    ai-session create-feature "$SUB_DIR" \
      --repo "$REPO" \
      --branch "<numbered-name>" \
      --work-dir "$WORK_DIR"
    ```

4.  Generate the content for `description.md` for the sub-story. The description must be self-contained and follow this structure:
    ```markdown
    ### User Story

    **Problem Description**
    <One paragraph: what is missing and why it matters. Reference the parent story if helpful.>

    **Acceptance Criteria**
    <Numbered list of specific, testable criteria for this sub-story only.>

    **Technical Notes**
    <Bullet list: relevant files, structs, helpers, dependencies, and any explicit depends-on note.>
    ```

5.  Write the generated content to `$SUB_DIR/description.md`. You MUST use `run_shell_command` with `printf` to write the file. For example:
    ```bash
    DESCRIPTION_CONTENT="your generated markdown content"
    printf '%s' "$DESCRIPTION_CONTENT" > "$SUB_DIR/description.md"
    ```

---

## 4. Confirm

After all directories are created, print a final summary table to stdout. The calling script will use this as the final output.

```
Sub-stories created under <FEATURE_BASE>:

  01  <numbered-name-1>   — <one-line summary>
  02  <numbered-name-2>   — <one-line summary>
  ...

Run `/session:start <sub-story-id>` to begin work on the first one.
```
