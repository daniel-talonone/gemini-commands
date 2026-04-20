---
created: 2026-02-25
modified: 2026-02-25
reviewed: 2026-02-25
name: Markdown Processing
description: Expert guidance for robustly reading, writing, and modifying Markdown files using standard command-line tools. Use this skill to avoid the fragility of simple text replacement and instead leverage the structure of Markdown documents.
allowed-tools: Bash, Read, Write, Edit, Grep, Glob
---

# Markdown Processing Skill

This skill provides expert guidance for performing structured, robust modifications to Markdown files using common command-line tools like `grep`, `sed`, and `awk`. It is designed to overcome the limitations and brittleness of context-unaware text replacement.

## Core Mandate

**Always prefer structured edits over simple replacement.** Analyze the Markdown structure (headings, lists, code blocks) to perform targeted and reliable changes. The `replace` tool should be a last resort when a structured approach is not feasible.

## Essential Patterns & Commands

### Finding Sections

The most robust way to edit a section is to identify its start and end line numbers.

```bash
# Find the line number of a heading (case-insensitive)
# The -n flag shows the line number. The -i flag makes it case-insensitive.
grep -n -i "^##* The Section Title" file.md

# Find the start and end line numbers of a section
# This requires two grep calls. One for the start, one for the next heading of the same or higher level.
START_LINE=$(grep -n -i "^## My Section" file.md | cut -d: -f1)
# Assuming the next section starts with '##' as well. Adjust the pattern as needed.
END_LINE=$(grep -n -A 9999 "## My Section" file.md | grep -m 1 -n "^## " | cut -d: -f1)
# END_LINE will be relative to the start line, so you may need to adjust.
# A better approach might be to find the line number of the *next* heading in the whole file.

# A more robust way to find the end of a section:
# 1. Find the line of the target section's heading.
# 2. Find the line of the NEXT heading of the same or lesser level.
# 3. The section content is between these two lines.

# Example: Get content of "## Usage" section
# Note: This is a conceptual example. The exact commands can be complex.
# It's often better to read the whole file and process it in the model's context if the logic gets too complex for one-liners.
```

### Replacing a Section's Content

Once you have the line numbers, you can use `sed` to replace the content of a whole section.

```bash
# (Conceptual) Replace content between two lines
# This is dangerous if not done carefully.
START_LINE=10
END_LINE=20
# sed is powerful but complex. A safer approach might be to read, modify in memory, and write back.
# For example:
# 1. read_file
# 2. Construct the new file content in the thought process
# 3. write_file to overwrite the original.
# This is often safer than complex sed/awk commands.
```

### Adding/Modifying List Items

#### Unordered Lists

```bash
# Add a new item to the end of a specific list.
# First, find the list. This can be tricky. A good way is to find a unique line in the list.
# Then, use sed to append.
# Example: Add "New Item" after the line containing "Existing Item".
sed -i '/Existing Item/a- New Item' file.md

# Replace a list item
sed -i 's/- Old Item/- New Item/' file.md
```

#### Ordered Lists

Editing ordered lists is harder because of numbering. It's often best to read the section, re-generate the list with correct numbering, and replace the entire list block.

### Modifying Code Blocks

Similar to sections, find the start and end of the code block using `grep` for the backticks with language specifier.

```bash
# Find a code block
grep -n "^```" file.md
```

Then, you can use `sed` or other tools to modify the content within that range. Again, reading, modifying, and writing back is often safer.

## Best Practices

1.  **Analyze First**: Always use `read_file` to understand the structure of the Markdown file before attempting any modification.
2.  **Identify Unique Anchors**: Find a unique heading, list item, or piece of text to anchor your change. Use `grep` to confirm your anchor is unique.
3.  **Prefer `read/modify/write`**: For complex changes (e.g., renumbering a list, complex section edits), the safest pattern is:
    a. `read_file` the entire file.
    b. Construct the **full, new content** of the file in your thought process.
    c. Use `write_file` to overwrite the old file with the new content. This avoids complex and error-prone `sed`/`awk` scripting.
4.  **Use `replace` Sparingly**: The `replace` tool should only be used for simple, unambiguous, and highly localized changes where you can provide significant, unique context. For example, fixing a typo in a single sentence. It is not suitable for replacing entire sections.
5.  **Validate Changes**: After using `write_file` or `sed -i`, use `read_file` again to confirm the change was applied as expected and didn't have unintended side-effects.

## Real-World Examples

### Example: Adding a new section to a README

**Goal**: Add a "## Contributing" section before the "## License" section in `README.md`.

**Workflow**:

1.  `read_file('README.md')` to get the current content.
2.  `grep -n "^## License" README.md` to find the line number where the new section should be inserted before. Let's say it's line 50.
3.  Construct the new content in memory.
4.  Use `sed` to insert the new section content at line 50.
    ```bash
    # This is complex. A read/modify/write approach is better.

    # Better approach:
    # 1. read_file('README.md') -> OLD_CONTENT
    # 2. In thought process, create NEW_CONTENT by splitting OLD_CONTENT at the "## License" heading, inserting the new section, and rejoining.
    # 3. write_file('README.md', NEW_CONTENT)
    ```

### Example: Updating a version number in a code example

**Goal**: In `docs/usage.md`, update the version number in a code block from `v1.2.0` to `v1.3.0`.

**Workflow**:

1.  `read_file('docs/usage.md')` to inspect the file.
2.  Decide if `replace` is safe. If the string `v1.2.0` is unique and the context is clear, `replace` can be used.
    ```python
    replace(
        file_path='docs/usage.md',
        old_string="...some unique context before...
    version: 'v1.2.0'
...some unique context after...",
        new_string="...some unique context before...
    version: 'v1.3.0'
...some unique context after...",
        instruction="..."
    )
    ```
3.  If `v1.2.0` is not unique, use `grep -n "v1.2.0"` to see all occurrences and find the correct line number and context to build a safe `replace` call or fall back to the `read/modify/write` pattern.
