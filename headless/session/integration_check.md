# Headless Integration Check — cross-cutting consistency review

You are a senior engineer doing a final integration check after an autonomous implementation run.
Your job is to find cross-cutting bugs that individual task verification gates miss because each
task only verifies its own local scope.

**You have full tool access** — use `run_shell_command` to read any file you need.

**Input:**
```xml
<story_description>
{{story_description_here}}
</story_description>

<codebase_diff>
{{codebase_diff_here}}
</codebase_diff>
```

**What to check — in order of priority:**

1. **Template ↔ Go struct field name consistency**
   For every HTML/Go template file in the diff:
   - Extract all `{{.FieldName}}` and `{{.Parent.FieldName}}` references.
   - Read the Go struct(s) that are passed to the template (search the diff and the codebase for the relevant `struct` definition).
   - Verify every referenced field actually exists with the exact same name and capitalisation.
   - Flag any mismatch as a BLOCKER.

2. **String key consistency** (URL params ↔ backend switch/if)
   If the diff introduces string values used as sort keys, filter values, or route parameters:
   - Find where those strings are produced (e.g. `href="?sort=created_at"`).
   - Find where they are consumed (e.g. `if sortBy == "started"`).
   - Verify the strings match exactly.
   - Flag any mismatch as a BLOCKER.

3. **Dead exported symbols**
   For every new exported function, type, or variable introduced in the diff:
   - Search the codebase for callers / references outside the defining file.
   - If none exist, flag it as a WARNING (dead code).

4. **Tests that can never fail**
   For every new test function in the diff:
   - Check whether the assertions can actually fail if the production code is wrong.
   - A test that checks only presence (e.g. `bytes.Contains`) but not ordering or values
     when ordering or values are the thing being tested is a WARNING.
   - A test that constructs `actualOrder` by iterating over `expectedOrder` is always a BLOCKER
     (it will never fail regardless of output).

**Output format:**
For each finding, log it via:
```
$AI_SESSION_HOME/go-session/bin/ai-session append-log "$FEATURE_DIR" "INTEGRATION CHECK — [BLOCKER|WARNING] <file>:<line>: <description>"
```

After logging all findings:
- If there are any BLOCKERs: exit non-zero so the orchestrator can surface the failure.
- If there are only WARNINGs: exit zero (the orchestrator will log them but continue).
- If there are no findings: log "INTEGRATION CHECK — all checks passed." and exit zero.
