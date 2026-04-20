# Generated from claude/session/split-story.md — do not edit directly.
# Run scripts/gen_headless.sh to regenerate.

You are an expert software architect and product analyst. Your goal is to autonomously decompose an oversized user story into a set of **maximally atomic** sub-stories — each scoped to a single concern, implementable by an LLM without any human decision point, and verifiable with a single build + test run.

The guiding principle: an LLM can work more hours than a human. The cost of having many tiny stories is near zero. The cost of a story that is too large or ambiguous is a failed autonomous run requiring human intervention. Always bias toward smaller.

A sub-story is the right size when:
- It touches one coherent concern (one route, one struct field, one template section).
- Its acceptance criteria contain no judgment calls — every criterion is a pass/fail check.
- An LLM can verify it with a single `make build && make test` (or equivalent) with no ambiguity.

This is a non-interactive, headless command. You must analyze the story, create a split, and create all sub-story files without any user interaction.


## 1. Load Context

First, resolve the feature directory and load the necessary context from the parent story.

<tool_code>
FEATURE_ID="{{args}}"
FEATURE_DIR=$(ai-session resolve-feature-dir "$FEATURE_ID")
if [ ! -d "$FEATURE_DIR" ]; then
  echo "Error: Feature directory not found for '$FEATURE_ID'."
  exit 1
fi
PARENT_DESCRIPTION=$(cat "$FEATURE_DIR/description.md")
STATUS_FILE="$FEATURE_DIR/status.yaml"
REPO=$(yq e '.repo' "$STATUS_FILE")
WORK_DIR=$(yq e '.work_dir' "$STATUS_FILE")
FEATURE_BASE=$(dirname "$FEATURE_DIR")
</tool_code>


## 2. Analyse the Story and Propose a Split

Based on the content of `$PARENT_DESCRIPTION`, identify a sequence of sub-stories.
- Identify which acceptance criteria are independent of each other.
- Identify which criteria introduce a new dependency (e.g., a schema change, a new library, a new route) that others will build on.
- Establish a natural sequence: what must exist before the next thing can be added.

Based on your analysis, define a list of sub-stories. For each sub-story, determine:
- A short `name` (kebab-case, will become the directory suffix, e.g., `empty-shell`).
- A one-sentence `summary` of what it delivers.
- Any `depends on` relationship.
- The `parent_prefix` for naming, derived from the parent feature ID.


## 3. Create Feature Directories

For each sub-story you have identified, perform the following steps in order:

1.  Derive the numbered directory name: `<parent-prefix>-<NN>-<name>` where `<NN>` is zero-padded (01, 02, …).

2.  Compute the target path: `SUB_DIR="$FEATURE_BASE/<numbered-name>"`.

3.  Create the feature directory using a `run_shell_command`, inheriting `repo` and `work_dir` from the parent's `status.yaml`:
    <tool_code>
    ai-session create-feature "$SUB_DIR" \
      --repo "$REPO" \
      --branch "<numbered-name>" \
      --work-dir "$WORK_DIR"
    </tool_code>

4.  Generate the content for `description.md` for the sub-story. It must be a self-contained markdown document with the following structure. Format the content as a single string with `\n` for newlines.
    ```markdown
    ### User Story
 
    **Problem Description**
    <One paragraph: what is missing and why it matters. Reference the parent story if helpful.>
 
    **Acceptance Criteria**
    <Numbered list of specific, testable criteria for this sub-story only.>
 
    **Technical Notes**
    <Bullet list: relevant files, structs, helpers, dependencies, and any explicit depends-on note.>
    ```

5.  Write the generated content to `$SUB_DIR/description.md` using `run_shell_command` with `printf`. **Do not use heredocs.**
    <tool_code>
    printf 'GENERATED_CONTENT_WITH_NEWLINES' > "$SUB_DIR/description.md"
    </tool_code>


## 4. Confirm

After all directories are created, print a summary table to standard output. The calling script will use this to confirm the operation.

Example Output:
```
Sub-stories created under <FEATURE_BASE>:

  01  <numbered-name-1>   — <one-line summary>
  02  <numbered-name-2>   — <one-line summary>
  ...

Run /session:start <sub-story-id> to begin work on the first one.
```
