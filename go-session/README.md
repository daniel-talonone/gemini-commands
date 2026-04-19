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
3. Otherwise — `~/.features/<org>/<repo>/<story-id>/`.

```bash
ai-session load-context <story-id>
```
Outputs all `.md`, `.yml`, and `.yaml` files in the feature directory as `<file name="...">content</file>` XML blocks, sorted alphabetically. Use this to load context into an LLM prompt.

### Log

```bash
ai-session append-log <feature-dir> <message>
```
Atomically appends a timestamped Markdown entry to `log.md`. Creates the file with a `# Work Log` header if it does not exist.

### PR Description

```bash
ai-session create-pr-description <feature-name>
```
Generates a PR description from feature context and writes it to `pr.md` via a headless LLM prompt.

Inputs: `description.md`, `plan.yml`, `log.md`, `status.yaml` (`work_dir`, `story_url`), git branch diff, and `.github/pull_request_template.md` from the repo (optional — skipped if absent). All inputs are injected into `headless/session/create-pr-description.md` and piped to `gemini --yolo`, which writes the result directly to `pr.md`. Re-running overwrites the file (idempotent). Sets `pipeline_step: pr-description-done` on success.

### Submit PR

```bash
ai-session submit-pr <feature-name>
```
Creates a GitHub PR for the feature's branch. The PR title is `feat: <branch-name>`. The PR body is read from `pr.md`. The base branch is detected automatically. If a PR already exists for the branch, the command exits with an error. On success, the PR URL is written to `status.yaml` and the `pipeline_step` is set to `pr-submitted`.


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
ai-session address-feedback <story-id> [--regular] [--docs] [--devops] [--remote]
```
Reads open findings per review type via `internal/review.ReadFindings` and pipes each to `gemini --yolo` using `headless/session/address-feedback.md`. This command now uses the same retry and verification engine as `implement`, running the project's verification command after each attempt and retrying on failure. Resolved findings are filtered out before the prompt is built. No flags → all three types are addressed. Types with no open findings are skipped automatically. The `--remote` flag fetches and addresses unresolved inline PR review threads from GitHub. It is mutually exclusive with the other flags and requires the `gh` CLI to be installed and authenticated.

#### Public API

-   `DiscoverTypes(featureDir string) ([]string, error)`
    Scans the feature directory for review files (`review*.yml`, `review*.yaml`) and returns a sorted list of their type names. An empty string (`""`) represents `review.yml`, `"docs"` represents `review-docs.yml`, and so on. Returns an empty slice if no review files are found.

-   `LoadByFilename(featureDir, filename string) ([]Finding, error)`
    Reads and validates all findings from a specific review file by its full name (e.g., `review-docs.yml`). Returns an empty slice if the file does not exist. It's the caller's responsibility to construct the correct filename or discover it first.

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
Starts a read-only HTTP dashboard for monitoring features at `http://localhost:1004`. The server scans `~/.features/` on every request, providing a real-time view of all tracked features.

**Main View:**
- **Filtering:** Supports `?repo=org/name` and `?status=running|idle|done`.
- **Sorting:** Supports `?sort=updated|started` with `&order=asc|desc`.
- **Quick-launch actions:** On macOS, each feature row displays icons to open the feature's working directory in Finder (📁), VSCode (`</>`), and the integrated terminal (⬛), provided a `work_dir` is set in its `status.yaml`.

**Detail View:**
- The header provides direct links to the feature's story (e.g., Shortcut or JIRA) and its associated pull request on GitHub.
- It also includes the same quick-launch actions (Finder, VSCode, Terminal) as the main view for easy access to the local development environment.

On macOS, the server also exposes `/action/terminal?path=<dir>`, `/action/finder?path=<dir>`, and `/action/vscode?path=<dir>` endpoints to open a directory in the respective application.

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
  github/               GitHub CLI interactions (PR review threads)
  log/                  log.md create and append (atomic writes)
  pr/                   pr.md create, read, and write (atomic writes)
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
