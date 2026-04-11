# ai-session CLI

A Go CLI that handles all deterministic file I/O for the `ai-session` workflow framework. It manages feature directories, plan files, and log entries so that LLM prompts never need raw `yq`, `sed`, or shell scripts for data mutations.

## Prerequisites

- Go 1.21+
- `golangci-lint` (for `make lint`)

## Build

```bash
make build      # produces bin/ai-session
make test       # runs all tests
make lint       # runs golangci-lint
make prcm       # build + test + lint (pre-commit)
```

## Subcommands

### Feature directories

```bash
ai-session create-feature <feature-dir>
  [--repo org/repo] [--branch name] [--work-dir /path]
```
Scaffolds a feature directory with `plan.yml`, `questions.yml`, `review.yml`, `log.md`, `pr.md`, and `status.yaml`. Derives `repo`, `branch`, and `work-dir` from git automatically. Never overwrites existing files.

```bash
ai-session resolve-feature-dir <story-id>
```
Prints the resolved path to a feature directory. Resolution order:
1. If `story-id` contains `/` or starts with `.`/`~` — used as-is.
2. If `.features/<story-id>/` exists in CWD — used (legacy).
3. Otherwise — `~/.ai-session/features/<org>/<repo>/<story-id>/`.

```bash
ai-session load-context <story-id>
```
Outputs all `.md`, `.yml`, and `.yaml` files in the feature directory as `<file name="...">content</file>` XML blocks, sorted alphabetically. Use this to load context into an LLM prompt.

### Log

```bash
ai-session append-log <feature-dir> <message>
```
Atomically appends a timestamped Markdown entry to `log.md`. Creates the file with a `# Work Log` header if it does not exist.

### Plan

```bash
ai-session plan-write <feature-dir>          # read from stdin; validates schema, then writes atomically
ai-session plan-list  <feature-dir>          # list slices
ai-session plan-list  <feature-dir> --slice <id>   # list tasks within a slice
ai-session plan-get   <feature-dir> --slice <id>          # full slice details
ai-session plan-get   <feature-dir> --slice <id> --task <id>  # full task body
ai-session plan-enrich-task <feature-dir> --slice <id> --task <id>  # update task body (stdin)
ai-session plan-split-task  <feature-dir> --slice <id> --task <id>  # replace task with subtasks (stdin YAML)
ai-session update-task  <feature-dir> <task-id>  --status <todo|in-progress|done>
ai-session update-slice <feature-dir> <slice-id> --status <todo|in-progress|done>
```

`plan-write` validates the plan against a strict schema (kebab-case IDs, non-empty descriptions, valid statuses) before writing. On success it also sets `pipeline_step: plan-done` in `status.yaml`.

`plan-enrich-task` and `plan-split-task` reject stdin containing `id:` or `status:` lines (injection guard) and only operate on tasks with status `todo`.

### Review

```bash
ai-session review <story-id> [--regular] [--docs] [--devops]
  [--strategy=branch|last-commit]
```
Runs a headless LLM code review via `gemini --yolo`. Fetches a diff, injects it into the matching headless prompt (`headless/session/review.md`, `review-docs.md`, or `review-devops.md`), and saves findings via `review-write`. The diff includes both tracked and untracked files.

- `--strategy=branch` (default): full branch diff vs `origin/<default-branch>`.
- `--strategy=last-commit`: uncommitted changes only (staged + unstaged + untracked vs HEAD).

No flags → all three types are reviewed. Flags are combinable.

```bash
ai-session review-write <feature-dir> --type <regular|docs|devops>
```
Validates and atomically writes review findings from stdin (YAML). The `internal/review` package is the single source of truth for filenames and format — callers never construct paths to `review*.yml` directly.

```bash
ai-session address-feedback <story-id> [--regular] [--docs] [--devops]
```
Reads open findings per review type via `internal/review.ReadFindings` and pipes each to `gemini --yolo` using `headless/session/address-feedback.md`. Resolved findings are filtered out before the prompt is built. No flags → all three types are addressed. Types with no open findings are skipped automatically.

### Orchestration

```bash
ai-session implement <story-id> [--max-retries 5] [--retry-delay 10s]
```
Headless LLM implementation loop. Reads `AGENTS.md` from the target project for the verification command and context file patterns, then iterates plan slices in dependency order — invoking `gemini --yolo` for each task, running the verification gate after each attempt, and retrying up to `--max-retries` times on failure. Rate-limit errors (HTTP 429, "quota exceeded") are retried on a separate budget (max 20) without consuming the main retry count. Sets `pipeline_step: implement-done` on completion.

```bash
ai-session start-plan <story-id>
```
Headless plan generation. Invokes `gemini --yolo` with a plan prompt, then runs `scripts/enrich_tasks.sh` to populate task bodies.

### Dashboard

```bash
ai-session serve [--port 1004]
```
Starts a read-only HTTP dashboard at `http://localhost:1004`. Scans `~/.ai-session/features/` on every request. Supports filters (`?repo=org/name`, `?status=running|idle|done`) and sorting (`?sort=updated|started&order=asc|desc`). On macOS, exposes `/action/terminal?path=<dir>` and `/action/finder?path=<dir>` endpoints to open a directory in Terminal.app or Finder.

## Package structure

```
cmd/ai-session/         CLI entry point (cobra); one file per subcommand
internal/
  commands/             Core business logic
    plan/               Plan YAML parsing, validation, format-preserving updates
    implement/          Headless LLM orchestration engine
    status/             status.yaml read/write
  dashboard/            Feature state derivation and directory scanning
  git/                  Git helper functions (remote URL, branch, work-dir, diff)
  log/                  log.md create and append (atomic writes)
  review/               review*.yml CRUD — Create, Load, Append, Write, UpdateStatus,
                        ReadFindings (open-only), AllTypes, TypeName
  server/               HTTP dashboard server and embedded HTML template
```

## Key design decisions

- **Atomic writes** — All file mutations write to a temp file and rename into place to prevent corruption on crash.
- **Format-preserving YAML** — Plan updates use the `yaml.Node` API to preserve original formatting and comments.
- **Strict plan validation** — `plan-write` rejects invalid YAML, missing required fields, non-kebab-case IDs, and unknown status values before any write occurs.
- **Idempotent feature creation** — `create-feature` skips files that already exist.
- **Self-documenting commands** — Every subcommand's `--help` output is sufficient to use it without reading the source.
