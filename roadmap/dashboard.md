# Dashboard & Orchestrator — Feature Roadmap

## Vision

`ai-session` evolves from a collection of prompts into a **multi-interface development workflow platform**. All interfaces share the same data layer — `~/.ai-session/features/{org}/{repo}/{feature}/` — and the same YAML schemas. Adding a new interface never requires changing the underlying data.

| Interface | Who uses it | How |
|---|---|---|
| Claude commands (`claude/session/`) | You, interactively | Chat prompt |
| Gemini commands (`gemini/session/`) | You, interactively | Chat prompt (generated from Claude) |
| Headless commands (`headless/session/`) | Orchestrator | `gemini --yolo -p` stateless pipe |
| **Go CLI (`ai-session`)** | **Prompts + Orchestrator** | **Shell commands** |
| Dashboard (`ai-session serve`) | You, visually | Browser |

The **Go CLI** is the stable interface between the AI and the filesystem. It handles all local file operations — reading feature directories, updating YAML, resolving paths, checking process state — so the AI never has to. The AI reasons and decides; the CLI executes and manages state.

The **dashboard** is the one piece that can see across all features, all repos, and all states simultaneously — something no individual LLM session can do. It is a subcommand of the same binary (`ai-session serve`), sharing the same internal packages as the CLI.

The **orchestrator** is the pipeline that runs session commands autonomously in the background, using the headless interface and the CLI for all file operations.

You declare a task (story ID + repo) and choose a mode — **automatic** (runs end-to-end without you) or **manual** (pauses at decision points). The dashboard surfaces only the moments that genuinely need you: answering open questions, making an architecture decision, triaging review findings.

Everything else — cloning, planning, enriching, implementing, reviewing, creating the PR — runs in the background.

---

## Core Concepts

### The Orchestrator

The `ai-session` Go service manages the full pipeline lifecycle. Given a story ID, repo URL, and mode, it runs:

```
new-feature → plan → enrich → implement → review (local) → address-feedback →
  commit-push → pr-description → submit-pr → [wait] →
  address-feedback --remote → commit-push → resolve-pr-threads → [repeat or done]
```

### Manual vs. Automatic Mode

The user declares the mode upfront, not the orchestrator at runtime. This is cleaner than trying to detect task complexity automatically.

**Automatic mode** — uses headless command variants. All decision gates are skipped, questions are auto-answered from the codebase where possible, and the pipeline runs unattended. Best for chore tasks: dependency bumps, metric tags, config changes.

**Manual mode** — uses the standard interactive commands. The pipeline pauses at architecture discussions, open questions, and review triage. The dashboard flags each pause and the user resumes from a pre-loaded terminal session.

The mode is stored in `status.yaml` so the dashboard can display it and the orchestrator knows how to resume after a pause.

### Stateless Command Execution

Session commands today are interactive chat sessions. The orchestrator replaces chat-based state passing with direct file piping — each step reads inputs from the filesystem and writes outputs back:

```bash
load_context_files.sh ~/.ai-session/features/org/repo/sc-1234 | gemini -p "$(cat headless/plan.md)"
```

### `AGENTS.md` Discovery with Global Fallback

To ensure the orchestrator has access to project-specific context even when operating on temporary clones (where `.gitignore`'d `AGENTS.md` files may not exist), `ai-session` uses a fallback discovery mechanism:
1.  **Local First**: It first looks for an `AGENTS.md` file in the root of the repository being worked on.
2.  **Global Fallback**: If not found locally, it checks for a corresponding file in a global directory: `~/.ai-session/agents/{git_org}/{repo_name}/AGENTS.md`.
3.  **Continue without**: If neither file exists, the session proceeds without the additional context.

This provides flexibility for developers to maintain local, un-versioned `AGENTS.md` files while ensuring that automated processes have a reliable way to access the same context.

### Centralized Feature Directories

Feature directories move from inside the repo (`.features/sc-1234/`) to a central location:

```
~/.ai-session/features/
  talon-one/payments-service/
    sc-1234/
      description.md   # includes repo (org/repo) and story_url
      architecture.md
      plan.yml
      questions.yml
      log.md
      review.yml
      pr.md
      status.yaml      # orchestrator metadata (new file)
  talon-one/user-service/
    sc-1301/
      ...
```

This decouples task knowledge from the repo lifecycle. Temporary clones can be deleted; the feature dir and all its decisions persist. Over time this becomes a searchable knowledge base of every task ever worked on.

`~/.ai-session/features/` is excluded from git via `.gitignore` — it is personal working state, not part of the repo.

### `status.yaml` Schema

A new file per feature, written and updated by the orchestrator. Stores only what cannot be derived from the other files — orchestrator metadata, not task progress (which lives in `plan.yml`, `questions.yml`, etc.).

```yaml
mode: auto                                    # auto | manual — set at invocation
story_url: https://app.shortcut.com/t1/story/1234  # direct link for dashboard
repo: talon-one/payments-service              # org/repo derived from git remote
branch: sc-1234                               # git branch created for this task
work_dir: /Users/you/code/payments-service   # absolute repo root on disk (git rev-parse --show-toplevel)
clone_path: /tmp/ai-session/sc-1234          # empty string if clone was cleaned up
pid: 12345                                    # orchestrator PID; 0 if not running
pipeline_step: new|plan|enrich|implement|review|pr|pr-submitted|pr-rejected|feedback|feedback-local-done|feedback-remote-done
error: "optional message on failure"          # last error message, for UI display
started_at: 2026-04-01T10:00:00Z
updated_at: 2026-04-01T14:22:00Z
pr_url: https://github.com/talon-one/payments-service/pull/123 # New field for PR URL
```

The `pipeline_step` indicates the current stage of the orchestration. When a step fails, it will be set to `<step>-failed` and the `error` field will be populated with a summary of the issue. The full error details will also be written to `log.md`.

**What is derived vs. stored:**
- `waiting_for` (blocked on questions/review/etc.) → **derived** from `questions.yml` and `review.yml` — never stored, avoids drift
- `is_running` → **derived** at read time via `kill(pid, 0)` — not stored
- Task progress → **derived** from `plan.yml` statuses — not stored
- Everything else above → **stored** (cannot be derived from other files)

### Human Intervention Detection

The dashboard derives task state entirely from file contents:

| Condition | Status | Needs you? |
|---|---|---|
| `plan.yml` missing | Not planned | ✅ Run `/session:plan` |
| `questions.yml` has `status: open` | Blocked on questions | ✅ Answer questions |
| No `architecture.md`, plan step blocked | Design decision needed | ✅ Architecture discussion |
| `plan.yml` has tasks `in-progress` | Implementing | — running |
| `review.yml` has `status: open` | Review findings to triage | ✅ Review decisions |
| `pr.md` has unresolved PR comments | Feedback to address | ✅ Address feedback |
| All tasks `done`, PR merged | Complete | — done |

In **automatic mode**, tasks only surface when the pipeline hits something it genuinely cannot resolve — an ambiguous requirement, a failing test it cannot fix, or a review finding that requires a judgment call.

### Go is the Glue, Not the AI

The Go CLI (`ai-session`) is the middleman between the AI and the filesystem. It handles everything the AI shouldn't have to think about: YAML parsing, path resolution, process state, atomic file writes, log formatting. The AI handles everything Go shouldn't: reasoning, planning, writing code, making decisions.

```
AI  →  reasons, plans, writes, decides
CLI →  reads files, updates YAML, resolves paths, checks process state
```

**CLI over MCP for local operations.** MCP servers shine for remote services — GitHub, Shortcut, Notion — where there is auth, pagination, and a long-lived connection to manage. For local filesystem operations, a CLI is simpler and more robust:

- No server process to run or register in settings files
- Debuggable: run the command yourself in the terminal
- Works with any LLM that has shell access — Claude, Gemini, any future tool
- The "API contract" is just `ai-session --help`
- Version-controlled alongside the prompts that use it

The current MCP integrations (GitHub, Shortcut, Notion) stay as MCPs — they're the right tool for remote services. `ai-session` owns the local layer.

**The prompts become LLM-agnostic by design.** Today the prompts depend on `yq` being the right version, `sed` behaving differently on macOS vs Linux, and complex shell escaping. Those problems move into Go, where they are testable and reliable. A prompt that calls `ai-session update-task sc-1234 add-user --status done` works identically on any machine, with any LLM.

---

## End-to-End Orchestration from the Dashboard

The dashboard will become the primary interface for initiating and monitoring the entire development lifecycle for a user story. This moves beyond single command execution to a fully orchestrated workflow managed by the `ai-session` Go service.

### "Add New Story" Workflow

A new view or modal in the dashboard will allow users to trigger a new session with the following inputs:
- **User Story URL**: The link to the story in Shortcut, Notion, etc.
- **Repository URL**: The full URL of the git repository.
- **Branch Name**: The name for the new git branch.
- **Mode**: `auto` or `manual`.
- **Local Work Directory** (Optional): An absolute path to an existing local clone of the repository. If provided, the orchestrator will use this directory instead of creating a temporary clone.

### Orchestration by `ai-session`

The `ai-session` Go service will directly manage the lifecycle, providing status updates to the dashboard at each step. **Each user story pipeline will run concurrently, so starting a new story does not block existing ones.** If any step fails or requires human input, the process pauses, and the dashboard will display a clear status message indicating the problem and the log output.

#### Managing Running Pipelines
- **Kill/Cancel Action**: Each feature row with a "running" status will display a "Cancel" icon. Clicking this will open a confirmation modal. If confirmed, the orchestrator will gracefully terminate the pipeline for that story.

#### `auto` Mode Pipeline

1.  **Prepare Workspace**:
    - If `Local Work Directory` is provided, use it.
    - Otherwise, clone the `Repository URL` into a temporary directory.
2.  **Sync Branch**: Checkout the main/master branch and pull the latest changes.
3.  **Create Branch**: Create and checkout a new branch with the specified `Branch Name`.
4.  **Initialize Session**: Execute `headless /session:new {user-story-id}`.
5.  **Create Plan**: Execute the `ai-session start-plan` command to generate the initial `plan.yml`.
6.  **Implement**: Execute `ai-session implement <story-id>` (Go orchestrator) to run all tasks.
7.  **Review**: Execute `headless /session:review` for an automated code review.
8.  **Address Feedback**: Automatically attempt to address any findings from the review step.
9.  **Generate PR Description**: Execute `ai-session create-pr-description <story-id>` to generate the PR description (`pr.md`).
10.  **Prepare for PR**: Pause, waiting for a human to give final approval before creating the pull request on GitHub.
#### `manual` Mode Pipeline

1.  **Prepare Workspace**: (Same as `auto` mode).
2.  **Sync Branch**: (Same as `auto` mode).
3.  **Create Branch**: (Same as `auto` mode).
4.  **Initialize Session**: Execute `headless /session:new {user-story-id}`.
5.  **Pause**: The process stops here. The dashboard will show the session is ready, and the user can proceed with manual steps (e.g., running `/session:plan` interactively) from their terminal.

### PR Approval Workflow

When the `auto` mode pipeline reaches the final "Prepare for PR" step, the dashboard will indicate that a PR is pending approval.
- An icon will appear next to the feature status.
- Clicking this icon opens a modal displaying the rendered `pr.md` content.
- The modal provides two modes:
  - **Preview Mode**: A read-only view of the PR description.
  - **Edit Mode**: A text area to edit the Markdown content, with "Save" and "Cancel" buttons. "Save" updates the local `pr.md` file.
- The modal also contains two actions:
  - **Approve**: Creates the pull request on GitHub using the current (and potentially edited) description from `pr.md`.
  - **Reject**: Aborts the PR creation, deletes the `pr.md` content, and updates the feature's `pipeline_step` to `pr-rejected`. The pipeline stops, awaiting user intervention.

---

## Components to Build

### 1. `scripts/resolve_feature_dir.sh`

Resolves the centralized feature dir path from the current git repo:

```bash
# From inside any repo:
resolve_feature_dir.sh sc-1234
# → /Users/you/.ai-session/features/talon-one/payments-service/sc-1234
```

Reads `git remote get-url origin`, extracts `org/repo`, returns the full central path. Used by all session commands to replace the hardcoded `.features/` prefix.

### 2. Headless Command Generator (`scripts/gen_headless.sh`)

Following the same pattern as `gen_gemini.sh`, a generator that produces headless variants of session commands from the Claude `.md` source files. The headless adapter prompt strips all interactive gates (architecture discussion, confirmation prompts) and replaces them with auto-defaults.

A **deny list** in the generator controls which commands are skipped — commands that are inherently interactive or user-facing and have no headless equivalent:

```bash
HEADLESS_DENY_LIST=("define" "start" "end" "get-familiar" "log-research" "migration" "verify-release")
```

Output goes to `claude/session/headless/*.md`, which are also committed and regenerated via the same workflow as Gemini commands. Single source of truth: `claude/session/*.md` → generates both `gemini/session/*.toml` and `claude/session/headless/*.md`.

Commands with headless variants:
- `new` — fetch story, scaffold dir, no user confirmation
- `plan` — skip architecture gate, auto-proceed through questions
- `review` / `review-docs` / `review-devops` — skip triage prompts, write findings directly
- `pr` — generate description and create PR without preview
- `address-feedback` — auto-apply clear fixes, flag ambiguous ones
- `checkpoint` / `summary` — already mostly stateless, minor changes

### 3. `/session:implement` Command ✅

Headless-only implementation command (`headless/session/implement.md`, hand-written). Reads `plan.yml`, runs an initial verification gate, then executes tasks slice by slice. After each task, runs the verification command and retries up to 5 times on failure. All exit conditions (success, initial failure, task failure, ambiguity, unmet dependency) are logged to `log.md`.

- **Headless mode**: runs all slices sequentially; stops and logs on any unrecoverable failure, marking the task `in-progress` for human triage
- **Interactive mode**: not yet implemented — future work

### Go CLI (`ai-session`)

A single Go binary with two modes: a CLI for file operations (used by prompts and orchestrator) and a dashboard server. One binary to compile, one thing in `$PATH`, one entry in `setup.sh`.

```bash
ai-session load-context sc-1234          # replaces scripts/load_context_files.sh
ai-session create-feature sc-1234        # replaces scripts/create_feature_dir.sh
ai-session create-pr-description sc-1234 # generate PR description from feature context and saves to pr.md
ai-session append-log sc-1234 "msg"      # replaces scripts/append_to_log.sh
ai-session update-task sc-1234 task-id --status done   # replaces yq one-liners
ai-session resolve-feature-dir sc-1234   # replaces scripts/resolve_feature_dir.sh
ai-session review sc-1234 [--regular] [--docs] [--devops] [--strategy=branch|last-commit]
ai-session review-write sc-1234 --type regular  # write findings from stdin (YAML)
ai-session address-feedback sc-1234 [--regular] [--docs] [--devops] [--remote]
ai-session add-plan-slice sc-1234 ...    # (to be created) appends a new slice to plan.yml
ai-session add-plan-task sc-1234 --slice-id ... # (to be created) appends a new task to a slice
ai-session serve                         # starts the dashboard web server
```

#### Go Architecture (Service-Oriented Refactoring)

The Go CLI is evolving towards a service-oriented architecture to act as a robust data access layer for all `ai-session` interfaces (Gemini, Claude, Headless, Dashboard).

Instead of monolithic commands or interfaces accessing files directly, functionality is being broken down into dedicated internal Go packages, or "services". Each service is responsible for a specific domain (e.g., `plan.yml`, `status.yaml`), encapsulating all logic, schema validation, and I/O for that domain's data.

This approach provides several key advantages:
- **Decoupling:** Interfaces (prompts, dashboard) are completely decoupled from the underlying data storage. They interact with the services, not the files.
- **Future-Proofing:** The data backend can be changed (e.g., from YAML files to a database) without requiring any changes to the client interfaces. The service contract remains the same.
- **Schema Enforcement:** Services act as a single gateway to the data, enforcing schema rules and ensuring data integrity across all clients.
- **Inter-Service Communication:** Services can safely interact with each other (e.g., the `plan` service can call a function on the `status` service) without needing to know about each other's underlying data representation.

The target architecture looks like this:
```
cmd/ai-session/
  main.go              ← Entry point, subcommand routing
internal/
  config/
    config.go          ← Handles ~/.ai-session/config.yml and agents discovery
  commands/
    status/
      status.go        ← Handles all reads and writes for status.yaml
    plan/
      plan.go          ← Handles all reads and writes for plan.yml
    review/
      review.go        ← (To be created) Handles review.yml
    log/
      log.go           ← (To be created) Handles log.md
    description/
      description.go   ← (To be created) Handles description.md
    pr/
      pr.go            ← (To be created) Handles pr.md
    (command files...)
  git/
    git.go             ← Common Git utility functions
  github/
    github.go          ← GitHub CLI interactions (PR review threads)
  server/
    server.go          ← HTTP server for `ai-session serve`
  ...
```

**Key libraries:**
- `gopkg.in/yaml.v3` — YAML parsing, maps directly to Go structs
- `github.com/fsnotify/fsnotify` — file system watching for real-time updates
- `net/http` + `html/template` — HTTP server and server-side rendering (no framework needed)
- `embed` — bundles the `frontend/dist/` folder into the binary at compile time

#### Data Flow

```
~/.ai-session/features/  →  scanner.go  →  state.go  →  HTTP handlers  →  frontend
                                                ↑
                                          fsnotify watch
                                          triggers SSE push
```

The scanner walks the feature directory tree, parses `status.yaml`, `plan.yml`, `questions.yml`, and `review.yml` for each feature, and passes raw structs to `state.go`. State derivation is pure Go logic — no file I/O, easily testable:

```go
type FeatureState struct {
    // From status.yaml
    Meta FeatureStatus
    // Parsed from plan.yml, questions.yml, review.yml
    Plan      []Slice
    Questions []Question
    Review    []Finding
    // Computed
    IsRunning   bool   // kill(pid, 0) check
    BlockedBy   string // "questions" | "architecture" | "review" | "pr-feedback" | ""
    DisplayStep string // human-readable current pipeline step
}
```

**Process liveness check** — `kill(pid, 0)` via `syscall` tells you if the orchestrator process is alive without sending a signal. If `pid > 0` and the call fails, the process has crashed.

#### Frontend Approach

**Go templates** for the initial version — no JS build step, no framework. Go renders HTML server-side; filtering is handled via URL query params (repo, status). No HTMX or SSE in MVP — page refresh is sufficient.

This can be replaced with a compiled React/Vue frontend later if richer interactions are needed — the Go API shape stays the same, only the `frontend/dist/` contents change.

#### MVP Iteration (v1)

**Feature list view — read-only:**
- Filter by repo (`?repo=talon-one/payments-service`) and status (`?status=running`).
- Basic pagination via query param (`?page=2`) to handle a large number of features.
- Per feature: story ID, repo, mode, current pipeline step (from `status.yaml`), whether running now (`kill(pid, 0)`), last completed task (last `done` task in `plan.yml`).
- Port: `1004` default, configurable via `--port` flag.

The dashboard is **read-only** — all writes happen through session commands and the orchestrator. The Go binary never modifies feature directories.

#### Next Iterations (post-MVP)

**v2 — Real-time updates:**
- Add `internal/watcher/watcher.go` (fsnotify wrapper) to watch `~/.ai-session/features/`
- Add SSE endpoint (`/events`) that pushes a reload signal on file changes
- Add minimal JS on the frontend to listen for SSE and refresh the page/component

**v3 — Quick-launch actions ✅ Done (partial):**
- Each feature row shows 📁/`</>`/⬛ icon buttons when `work_dir` is set in `status.yaml`
- Folder → `file:///work_dir` (Finder, no server needed). VSCode → `vscode://file/work_dir` (no server needed). Terminal → `GET /action/terminal?path=work_dir` runs `open -a Terminal <path>`.
- `work_dir` (repo root on disk) added to `status.yaml`, populated at creation time by `ai-session create-feature` and `scripts/create_feature_dir.sh` via `git rev-parse --show-toplevel`.
- Note: `html/template` sanitizes non-http/https URI schemes — a `safeURL` FuncMap entry is required for `file://` and `vscode://` links; the terminal endpoint URL uses built-in `urlquery`.
- **Still pending in v3:** An "Add New Story" form/modal and a "Resume" button for blocked tasks. The detailed workflow for adding a new story is described in the **End-to-End Orchestration from the Dashboard** section.

**v4 — Task detail view:**
- Clicking a feature row opens a dedicated detail view.
- **Header**: Contains links to the Shortcut/Notion story and the GitHub PR (if they exist), along with creation/update timestamps from `status.yaml`.
- **Body**: Renders the full `description.md` and `log.md` as GitHub-flavored Markdown.
- **Sidebar**: Shows open items from `questions.yml` and `review.yml` as actionable cards.

**v5 — Configuration View:**
- A new page in the dashboard (e.g., `/config`) for managing the global `~/.ai-session/config.yml`.
- The view will display the current configuration, including a list of known repositories.
- For each repository, the user can see and edit its URL and default local `work_dir`.
- The user can add new repository configurations or remove existing ones.
- This requires new backend API endpoints (e.g., `GET /api/config`, `POST /api/config`) and leveraging the `internal/config` service to handle file I/O.

---

## Prerequisites & Dependencies

| Prerequisite | Status | Notes |
|---|---|---|
| `enrich_tasks.sh` | ✅ Done | Background plan enrichment. Three-mode output contract: `ENRICH:` (clarify with FILE/FUNCTION/code), `SPLIT:` (replace with N atomic subtasks via `plan-split-task`), `SKIP:` (already detailed enough — LLM-decided, not heuristic). Enricher receives slice context (sibling task summaries) to avoid redundant splits. |
| `gen_gemini.sh` + adapter prompt | ✅ Done | Stateless command generation pattern (reused by headless generator) |
| `plan.yml` nested slice/task schema | ✅ Done | Required for implement step |
| Centralized feature dirs + `resolve_feature_dir.sh` | ✅ Done | Backward-compatible three-level resolution; all session commands updated |
| `gen_headless.sh` + headless commands | ✅ Done | 9 commands generated; `verify-release`, `plan` hand-written/denied; all hand-fixed and reviewed. Run with `gemini --yolo -p` (requires agentic mode for `run_shell_command`). Heredoc syntax prohibited in adapter prompt (Gemini CLI parser rejects `<< 'EOF'`). Deny list: `define`, `start`, `end`, `get-familiar`, `log-research`, `migration`, `plan`, `checkpoint`, `implement`, `review`, `review-docs`, `review-devops`, `address-feedback`. The `review*` and `address-feedback` headless prompts are hand-written — they receive injected `{{diff_here}}` / `{{findings_here}}` / `{{feature_dir_here}}` from the Go orchestrator rather than fetching files themselves. |
| `/session:implement` command + `ai-session implement` | ✅ Done | Two layers: (1) `headless/session/implement.md` — LLM-driven headless prompt (hand-written). (2) `ai-session implement <story-id>` — Go orchestrator that replaces the headless prompt with a deterministic loop: reads `AGENTS.md` for the verification command, runs an initial gate, iterates slices with dependency checks, invokes a new `headless/session/execute_task.md` prompt per task via Gemini stdin, retries up to 5 times on failure, updates task/slice statuses atomically, sets `pipeline_step: implement-done` on completion. `plan.Plan`/`Slice`/`Task` types exported from `internal/commands/plan`. |
| Go CLI (`ai-session`) — CLI subcommands | ✅ Done | 15 subcommands: `create-feature`, `resolve-feature-dir`, `append-log`, `update-task`, `update-slice`, `plan-list`, `plan-get`, `plan-write`, `plan-enrich-task`, `plan-split-task`, `load-context`, `implement`, `review`, `review-write`, `address-feedback`. Cobra CLI, yaml.Node API, testify. Makefile with build/test/lint. `plan-write` validates schema + writes atomically; `plan-enrich-task` updates single task field with injection guard and status lock; `plan-split-task` replaces one todo task with N atomic subtasks; `load-context` outputs feature dir files as XML blocks; `implement` is the Go orchestrator for the implementation phase (see `/session:implement` row). `review` fetches a diff and pipes it to a headless LLM prompt — supports `--strategy=branch` (full branch diff vs origin, default) or `--strategy=last-commit` (uncommitted changes vs HEAD); both strategies include untracked files. `review-write` validates and atomically writes review findings from stdin. `ai-session address-feedback sc-1234 [--regular] [--docs] [--devops] [--remote]`
Reads open findings per review type via `internal/review` and pipes each to `gemini --yolo` using the headless address-feedback prompt. This command now uses the same retry and verification engine as `implement`, running the project's verification command after each attempt and retrying on failure. The `--remote` flag also fetches and addresses unresolved inline PR review threads from GitHub. |
| `address-feedback --remote` | ✅ Done | Fetches unresolved PR review threads via `gh pr view --json reviewThreads` (no positional branch arg; `cmd.Dir` used for context) and injects them into a new `headless/session/address-feedback-remote.md` prompt. `internal/github` package (`GetUnresolvedReviewThreads`) encapsulates all GitHub CLI interactions; file-level comments omit the `:0` line suffix. Two new `pipeline_step` values: `feedback-local-done` and `feedback-remote-done`. Template uses `{{feature_dir}}` placeholder, replaced by the orchestrator at runtime. |
| Go CLI (`ai-session serve`) — dashboard | ✅ Done | MVP read-only dashboard. Scans `~/.ai-session/features/` on each request. Feature list with repo/status filters. Per-feature: story ID, repo, mode, pipeline step, running indicator (`kill(pid,0)`), last done task. Quick-launch icons (📁 Finder / `</>` VSCode / ⬛ Terminal) when `work_dir` set. `GET /action/terminal` endpoint. Go templates + `go:embed`. Port 1004 default, `--port` flag. Graceful shutdown. `status.yaml` scaffolded by `create-feature` with `repo`, `branch`, `work_dir`, `started_at`, `updated_at` from git. |

### Scripts → Go CLI Migration Path

The shell scripts in `scripts/` are not rewritten all at once. The Go CLI grows incrementally alongside them:

1. Build `ai-session serve` (dashboard) first — immediate value, no prompt changes needed
2. Add CLI subcommands one by one, starting with the most painful (`yq` mutations → `update-task`)
3. Update prompts to call `$AI_SESSION_HOME/bin/ai-session <subcommand>` as each script is replaced
4. Deprecate `scripts/` once the CLI covers all operations

The shell scripts remain as fallback during the transition. No big-bang rewrite. The prompts already use the `$AI_SESSION_HOME/scripts/` convention, so swapping to `$AI_SESSION_HOME/bin/ai-session` is mechanical.

---

## Suggested Build Order

1. **Centralized feature dirs** — `resolve_feature_dir.sh` + update session commands. Low risk, high impact. Enables the knowledge database immediately even before the orchestrator exists.
2. **`gen_headless.sh`** — same pattern as `gen_gemini.sh`, single source of truth for headless variants. Relatively cheap to build once the generator pattern is understood.
3. **`/session:implement`** — the biggest missing piece. Start with interactive mode (slice-by-slice confirmation); headless variant is generated by `gen_headless.sh`.
4. **Dashboard** — the UI layer. Start with a generated static HTML file; evolve to a persistent local web server if needed.

---

## Known Bugs

- ~~**`enrich_tasks.sh` corrupts `status` fields**~~ — **Fixed.** The enricher now uses `ai-session plan-enrich-task` to update only the `task:` field of one todo task at a time. An injection guard rejects any LLM output containing `id:` or `status:` lines; tasks with `in-progress` or `done` status are skipped entirely. Full-file LLM overwrite of `plan.yml` is no longer possible.

- **`scripts/gen_gemini.sh` has leftover debug `echo` statements** — added during bash 3.2 associative-array fix investigation, never removed. Lines log `DEBUG: CLAUDE_DIR`, `CHECKSUMS_FILE`, `UPDATED_CHECKSUMS`, `FORCE`, `md files`, and per-file `stored`/`current`/`toml`/`description` values. Safe to delete; grep for `echo "DEBUG:` in the file.

- ~~**`enrich_tasks.sh` fails silently when launched via `nohup`**~~ — **Not a bug.** `go-session/bin/` is added to PATH in `.zshenv`, which is sourced even for non-interactive shells. The issue was a stale terminal session that predated the `.zshenv` update; verified 2026-04-04.

- **`os.Exit(1)` in `RunE` handlers** — Several cmd files call `os.Exit(1)` inside `RunE` instead of returning the error. This bypasses Cobra's error handling and makes the exit code untestable. Should return `fmt.Errorf(...)` instead and let Cobra print the error. `cmd_review_write.go` has been fixed; remaining files not yet audited. Low priority; no user-visible behavior change.

- **`plan-write` loses YAML formatting** — `plan-write` preserves the original bytes on write, but `updateStatusPipelineStep` (triggered as a side-effect) reads and re-marshals `status.yaml` via `yaml.Marshal`, which normalises formatting (removes single-quote wrapping from empty strings). Cosmetic only — YAML semantics are preserved.

- **`repo` slug edge cases in `ParseOrgRepo`** — repos with more than one `/` in the path (e.g. GitHub Enterprise with a subdirectory prefix) may not parse correctly. `strings.Index` approach handles standard `org/repo` only.

---

## Open Questions

- **`/session:implement` granularity**: ~~Resolved~~ — headless mode retries up to 5 times per task before stopping. Interactive mode (slice-level confirmation) is not yet implemented.
- **Story link format**: `description.md` stores the raw story ID (`sc-1234`) or full URL? Full URL is unambiguous and directly linkable in the dashboard. `/session:new` already fetches from Shortcut/Notion so it has the URL available.
- **Resume signal**: When the orchestrator pauses waiting for intervention, how does "resume" work? Options: a sentinel file the user creates (`touch .resume`), a simple CLI command (`ai-session resume sc-1234`), or the dashboard triggers it via a button.
- **Dashboard tech**: Generated static HTML (lowest friction) vs. a persistent local web server (enables file-watch, push updates). Start static, evolve if needed.
- **Multi-user**: Personal-only for now. Shared team use would require the central feature dir to live in a shared location (network path or a dedicated git repo).
- **`internal/yaml` wrapper package**: As YAML parsing spreads across `internal/commands/` (yaml.Node for writes, struct tags for reads) and `internal/dashboard/` (struct tags for reading status.yaml and plan.yml), there may be value in a shared `internal/yaml` package to avoid duplication. Defer until a third consumer appears.
- **`AGENTS.md` loading in `ai-session load-context`**: The initial design included an `--agents-file` flag to append `AGENTS.md` content to the context output. This was deferred — there is a non-trivial situation around how `AGENTS.md` is discovered, located, and whether it belongs in the CLI layer or stays the caller's responsibility. Needs dedicated exploration before implementing.
- **`/session:amend` — mid-plan scope changes**: When the user changes their mind after planning has started (scope cut, new constraint, revised approach), there is currently no systematic way to propagate the change. The result is drift between `description.md`, `questions.yml`, plan task descriptions, and any already-written files. A `/session:amend` command could address this: the user describes what changed, and the command updates all affected artifacts atomically — revising the description, closing/reopening questions, rewriting stale task bodies, and flagging any already-implemented tasks that may need revisiting. Key design questions: how to detect which tasks are affected, whether to re-validate already-`done` tasks, and whether this is interactive-only or also has a headless variant.

- **Unifying all session interfaces in the dashboard**: The goal is for the dashboard to be the single management UI for all session interfaces — Claude commands, Gemini commands, headless/orchestrator, and manual sessions. All interfaces already share the same `status.yaml` data layer. The missing piece is a consistent `pipeline_step` vocabulary and ensuring every interface writes `status.yaml` (currently only `plan-write` and the Go orchestrator steps do so). Manual sessions (no orchestrator) leave `pipeline_step` empty after `plan-done` — the dashboard shows them as idle. To fully unify: Claude and Gemini `/session:implement`, `/session:review`, and `/session:pr` commands should update `pipeline_step` at start/end of each step. This is cheap to add once the vocabulary is agreed.

- **`pipeline_step` vocabulary not fully standardised**: Currently defined: `plan-done` (by `plan-write`), `implement-done` (by `ai-session implement`), `feedback-local-done` (by `ai-session address-feedback`), `feedback-remote-done` (by `ai-session address-feedback --remote`), `pr-submitted` (by `ai-session submit-pr`). No standard for `new`, `enrich`, `review-done`, `pr-done`, or `done`. Should define a canonical enum before adding more writers.
