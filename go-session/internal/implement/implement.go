package implement

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/daniel-talonone/gemini-commands/internal/description"
	"github.com/daniel-talonone/gemini-commands/internal/gemini"
	"github.com/daniel-talonone/gemini-commands/internal/log"
	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/daniel-talonone/gemini-commands/internal/status"
	"gopkg.in/yaml.v3"
)

// runShellAndCaptureOutput executes a shell command and captures its stdout/stderr.
func runShellAndCaptureOutput(cmdStr string) (string, error) {
	var output bytes.Buffer
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	return output.String(), err
}

type Strategy interface {
	ExecuteSlice(ctx SliceContext) error
}

type SliceContext struct {
	FeatureDir      string
	WorkDir         string
	AISessionHome   string
	Story           string
	Architecture    string
	Slice           plan.Slice
	VerificationCmd string
	MaxRetries      int
	RetryDelay      time.Duration
	ContextPattern  []string
	Logger          *slog.Logger
}

type PerTaskStrategy struct{}

func (s *PerTaskStrategy) ExecuteSlice(ctx SliceContext) error {
	for j := range ctx.Slice.Tasks {
		t := &ctx.Slice.Tasks[j]
		if t.Status == "done" {
			continue
		}

		resuming := t.Status == "in-progress"
		if resuming {
			appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Resuming task: %s (was in-progress from a prior run)", t.ID))
		} else {
			appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Starting task: %s...", t.ID))
		}
		ctx.Logger.Info("Starting task", "slice", ctx.Slice.ID, "task", t.ID, "resuming", resuming)

		if err := plan.UpdateTask(ctx.FeatureDir, t.ID, "in-progress"); err != nil {
			return fmt.Errorf("updating task %s to in-progress: %w", t.ID, err)
		}

		if err := executeTaskWithRetry(ctx.Logger, ctx.FeatureDir, ctx.AISessionHome, ctx.WorkDir, ctx.Story, ctx.Architecture, ctx.Slice.Description, t.Task, ctx.VerificationCmd, ctx.MaxRetries, ctx.RetryDelay, ctx.ContextPattern); err != nil {
			appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Task %s FAILED after all retries: %v", t.ID, err))
			ctx.Logger.Error("Task failed", "task", t.ID, "error", err)
			return fmt.Errorf("task %s in slice %s failed: %w", t.ID, ctx.Slice.ID, err)
		}

		if err := plan.UpdateTask(ctx.FeatureDir, t.ID, "done"); err != nil {
			return fmt.Errorf("updating task %s to done: %w", t.ID, err)
		}
		t.Status = "done"
		appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Task %s completed successfully.", t.ID))
		ctx.Logger.Info("Task completed", "task", t.ID)
	}
	return nil
}

type PerSliceStrategy struct{}

func (s *PerSliceStrategy) ExecuteSlice(ctx SliceContext) error {
	promptPath := filepath.Join(ctx.AISessionHome, "headless", "session", "execute_slice.md")
	promptTemplate, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("reading execute_slice.md: %w", err)
	}

	const maxRateLimitRetries = 20
	attempt := 1
	rateLimitRetries := 0
	var lastError string

	for attempt <= ctx.MaxRetries {
		if attempt > 1 && ctx.RetryDelay > 0 {
			ctx.Logger.Info("Waiting before retry", "delay", ctx.RetryDelay, "slice", ctx.Slice.ID, "attempt", attempt)
			time.Sleep(ctx.RetryDelay)
		}

		// Re-read tasks from disk on each attempt so statuses reflect prior LLM updates.
		currentPlan, err := plan.LoadPlan(ctx.FeatureDir)
		if err != nil {
			return fmt.Errorf("reloading plan for slice %s: %w", ctx.Slice.ID, err)
		}
		currentSlice, found := currentPlan.FindSlice(ctx.Slice.ID)
		if !found {
			return fmt.Errorf("slice %s not found in plan", ctx.Slice.ID)
		}

		// Serialize tasks as valid YAML (handles multi-line descriptions safely).
		tasksBytes, err := yaml.Marshal(currentSlice.Tasks)
		if err != nil {
			return fmt.Errorf("serializing tasks for slice %s: %w", ctx.Slice.ID, err)
		}

		// Re-compute diff on each attempt to reflect latest codebase state.
		changesSoFar := getSourceFilesDiff(ctx.WorkDir, ctx.ContextPattern)

		promptContent := strings.ReplaceAll(string(promptTemplate), "{{story_description_here}}", ctx.Story)
		promptContent = strings.ReplaceAll(promptContent, "{{architecture_description_here}}", ctx.Architecture)
		promptContent = strings.ReplaceAll(promptContent, "{{slice_description_here}}", currentSlice.Description)
		promptContent = strings.ReplaceAll(promptContent, "{{tasks_here}}", string(tasksBytes))
		promptContent = strings.ReplaceAll(promptContent, "{{changes_so_far_here}}", changesSoFar)
		promptContent = strings.ReplaceAll(promptContent, "{{verification_command_here}}", ctx.VerificationCmd)
		promptContent = strings.ReplaceAll(promptContent, "{{feature_dir_here}}", ctx.FeatureDir)

		if lastError != "" {
			promptContent = strings.ReplaceAll(promptContent, "{{error_message_here}}", lastError)
			promptContent = strings.ReplaceAll(promptContent, "{{#if error_message}}", "")
			promptContent = strings.ReplaceAll(promptContent, "{{/if}}", "")
		} else {
			promptContent = errorBlockRe.ReplaceAllString(promptContent, "")
		}

		var geminiOutput bytes.Buffer
		if os.Getenv("IN_TEST_MODE") != "true" {
			appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Invoking LLM for slice execution: %s (attempt %d)", ctx.Slice.ID, attempt))
			ctx.Logger.Info("Invoking Gemini for slice", "slice", ctx.Slice.ID, "attempt", attempt)
			if err := gemini.RunYolo(strings.NewReader(promptContent), io.MultiWriter(os.Stdout, &geminiOutput), io.MultiWriter(os.Stderr, &geminiOutput)); err != nil {
				appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Gemini exited with error (slice %s, attempt %d): %v", ctx.Slice.ID, attempt, err))
				if isRateLimitError(geminiOutput.String()) {
					rateLimitRetries++
					if rateLimitRetries >= maxRateLimitRetries {
						return fmt.Errorf("exceeded rate-limit retry budget for slice %s", ctx.Slice.ID)
					}
					continue // rate-limit retries don't consume the attempt budget
				}
				// Non-rate-limit Gemini error: still check gates — file changes may have landed.
			}
		} else {
			ctx.Logger.Info("Skipping Gemini invocation (test mode)", "slice", ctx.Slice.ID)
		}

		// Gate 1: task completeness — all tasks must be marked done by the LLM.
		reloadedPlan, err := plan.LoadPlan(ctx.FeatureDir)
		if err != nil {
			return fmt.Errorf("reloading plan after slice %s execution: %w", ctx.Slice.ID, err)
		}
		reloadedSlice, found := reloadedPlan.FindSlice(ctx.Slice.ID)
		if !found {
			return fmt.Errorf("slice %s not found in reloaded plan", ctx.Slice.ID)
		}
		var incomplete []string
		for _, t := range reloadedSlice.Tasks {
			if t.Status != "done" {
				incomplete = append(incomplete, fmt.Sprintf("%s (status: %s)", t.ID, t.Status))
			}
		}
		if len(incomplete) > 0 {
			lastError = fmt.Sprintf("not all tasks marked done: %s", strings.Join(incomplete, ", "))
			appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Gate 1 failed for slice %s (attempt %d): %s", ctx.Slice.ID, attempt, lastError))
			attempt++
			continue
		}

		// Gate 2: verification.
		verificationOutput, verifyErr := runShellAndCaptureOutput(ctx.VerificationCmd)
		if verifyErr != nil {
			lastError = fmt.Sprintf("verification failed: %v\nOutput:\n%s", verifyErr, verificationOutput)
			appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Gate 2 failed for slice %s (attempt %d): %v", ctx.Slice.ID, attempt, verifyErr))
			attempt++
			continue
		}

		appendLog(ctx.Logger, ctx.FeatureDir, fmt.Sprintf("Slice %s: all gates passed (attempt %d).", ctx.Slice.ID, attempt))
		ctx.Logger.Info("All gates passed", "slice", ctx.Slice.ID, "attempt", attempt)
		return nil
	}

	return fmt.Errorf("slice %s failed after %d attempts: %s", ctx.Slice.ID, ctx.MaxRetries, lastError)
}

// errorBlockRe strips the entire {{#if error_message}}...{{/if}} block (including
// its XML wrapper) when there is no error to inject.
var errorBlockRe = regexp.MustCompile(`(?s)\{\{#if error_message\}\}.*?\{\{/if\}\}`)

// agentsVerifyRe extracts the verification command from an AGENTS.md "## Verification" section.
var agentsVerifyRe = regexp.MustCompile(`(?m)^## Verification\r?\nRun: (.+)$`)

// agentsContextFilesRe extracts the source file glob pattern from an AGENTS.md
// "## Context files\nPattern: <globs>" section.
var agentsContextFilesRe = regexp.MustCompile(`(?m)^## Context files\r?\nPattern: (.+)$`)

// defaultContextExcludes are git pathspec exclusions applied when no explicit
// "## Context files" pattern is configured in AGENTS.md. They strip known
// non-source files (config, docs, lock files, binaries) that produce formatting
// noise without containing any structural information the LLM needs.
var defaultContextExcludes = []string{
	":!*.yaml", ":!*.yml", ":!*.json", ":!*.toml",
	":!*.md", ":!*.txt", ":!*.lock", ":!*.sum",
	":!*.svg", ":!*.png", ":!*.jpg", ":!*.gif", ":!*.ico",
}

// appendLog writes a timestamped entry to log.md. Logging is best-effort — a
// failure is reported via the logger but never stops the orchestration.
func appendLog(logger *slog.Logger, featureDir, msg string) {
	if err := log.AppendLog(featureDir, msg); err != nil {
		logger.Warn("failed to append log entry", "error", err)
	}
}

// Run orchestrates the full implementation loop for a feature: reads the plan,
// runs the verification gate, iterates slices/tasks, invokes the LLM per task,
// and marks statuses as it progresses.
//
// workDir must be the target project root (the directory containing AGENTS.md).
// aiSessionHome must be the resolved AI_SESSION_HOME path.
// maxRetries is the maximum number of LLM+verification attempts per task.
// retryDelay is the pause between attempts (helps avoid LLM rate limits).
func Run(logger *slog.Logger, featureID, featureDir, workDir, aiSessionHome string, maxRetries int, retryDelay time.Duration, strategy Strategy) (err error) {
	// Set the initial pipeline_step. If the previous run left "implement-failed"
	// (i.e. a human manually intervened and is restarting), use "implement-restarted"
	// so the dashboard reflects the context. Any other state is treated as a first run.
	currentStep, _ := status.ReadStep(featureDir)
	initialStep := "implement"
	if currentStep == "implement-failed" {
		initialStep = "implement-restarted"
	}
	if writeErr := status.Write(featureDir, initialStep, "", ""); writeErr != nil {
		logger.Warn("failed to write initial implement status", "error", writeErr)
	}

	// On any failure, mark the run as failed so the dashboard and the next invocation
	// can distinguish between "in progress", "failed", and "done".
	defer func() {
		if err != nil {
			if writeErr := status.Write(featureDir, "implement-failed", "", ""); writeErr != nil {
				logger.Warn("failed to write implement-failed status", "error", writeErr)
			}
		}
	}()

	appendLog(logger, featureDir, fmt.Sprintf("--- Starting implementation orchestration for feature: %s ---", featureID))
	logger.Info("Starting implementation orchestration", "feature_id", featureID)

	// Extract verification command from the project's AGENTS.md.
	verificationCmd, err := extractVerificationCommand(workDir)
	if err != nil {
		return fmt.Errorf("extracting verification command: %w", err)
	}

	// Optional source-file filter for the codebase context diff (see AGENTS.md
	// "## Context files" section). Falls back to defaultContextExcludes if absent.
	contextPattern := extractContextPattern(workDir)
	appendLog(logger, featureDir, fmt.Sprintf("Using verification command: %s", verificationCmd))
	logger.Info("Using verification command", "command", verificationCmd)

	// Initial verification gate — codebase must be passing before we start.
	appendLog(logger, featureDir, "Running initial verification gate...")
	logger.Info("Running initial verification gate")
	if err := runShell(verificationCmd); err != nil {
		msg := fmt.Sprintf("Initial verification failed — codebase must be in a passing state to begin: %v", err)
		appendLog(logger, featureDir, msg)
		return fmt.Errorf("%s", msg)
	}
	appendLog(logger, featureDir, "Initial verification gate passed.")
	logger.Info("Initial verification gate passed.")

	// Read story description for prompt context.
	storyDescription, err := description.LoadDescription(featureDir)
	if err != nil {
		return fmt.Errorf("reading description.md: %w", err)
	}

	// Load architecture if present — gives the LLM design constraints and pattern refs.
	architectureDescription, err := description.LoadArchitecture(featureDir)
	if err != nil {
		return fmt.Errorf("loading architecture: %w", err)
	}

	// Load plan.
	p, err := plan.LoadPlan(featureDir)
	if err != nil {
		return fmt.Errorf("loading plan: %w", err)
	}

	// Main loop: process slices in dependency order using a progress-made pattern
	// to handle slices whose dependencies become satisfied mid-run.
	for {
		progressMade := false
		allDone := true

		for i := range p {
			s := &p[i]
			if s.Status == "done" {
				continue
			}
			allDone = false

			// Check all depends_on slices are done (consult in-memory Plan only).
			if !depsMet(p, s.DependsOn) {
				continue
			}

			appendLog(logger, featureDir, fmt.Sprintf("Starting slice: %s", s.ID))
			logger.Info("Starting slice", "id", s.ID)

			if err := plan.UpdateSlice(featureDir, s.ID, "in-progress"); err != nil {
				return fmt.Errorf("updating slice %s to in-progress: %w", s.ID, err)
			}

			sliceCtx := SliceContext{
				Logger:          logger,
				FeatureDir:      featureDir,
				AISessionHome:   aiSessionHome,
				WorkDir:         workDir,
				Story:           storyDescription,
				Architecture:    architectureDescription,
				VerificationCmd: verificationCmd,
				MaxRetries:      maxRetries,
				RetryDelay:      retryDelay,
				ContextPattern:  contextPattern,
				Slice:           *s,
			}

			if err := strategy.ExecuteSlice(sliceCtx); err != nil {
				return fmt.Errorf("slice %s failed: %w", s.ID, err)
			}

			if err := plan.UpdateSlice(featureDir, s.ID, "done"); err != nil {
				return fmt.Errorf("updating slice %s to done: %w", s.ID, err)
			}
			s.Status = "done"
			appendLog(logger, featureDir, fmt.Sprintf("Slice %s completed.", s.ID))
			logger.Info("Slice completed", "id", s.ID)
			progressMade = true
		}

		if allDone {
			break
		}
		if !progressMade {
			return fmt.Errorf("cannot make further progress — check for unsatisfiable or circular dependencies")
		}
	}

	// Final integration check: catches cross-cutting issues (template↔struct mismatches,
	// sort key inconsistencies, dead code, always-passing tests) that individual task
	// verification gates miss because each task only validates its own local scope.
	appendLog(logger, featureDir, "Running final integration check...")
	logger.Info("Running integration check")
	if err := runIntegrationCheck(logger, aiSessionHome, storyDescription, workDir, contextPattern); err != nil {
		appendLog(logger, featureDir, fmt.Sprintf("Integration check found blockers: %v", err))
		return fmt.Errorf("integration check failed: %w", err)
	}
	appendLog(logger, featureDir, "Integration check passed.")

	if err := status.Write(featureDir, "implement-done", "", ""); err != nil {
		return fmt.Errorf("updating status to implement-done: %w", err)
	}
	appendLog(logger, featureDir, "--- IMPLEMENT COMPLETE ---")
	logger.Info("IMPLEMENT COMPLETE")
	return nil
}

// extractContextPattern reads the optional "## Context files\nPattern: <globs>" section
// from AGENTS.md and returns the space-separated glob list. Returns nil when the section
// is absent so callers can fall back to defaultContextExcludes.
func extractContextPattern(dir string) []string {
	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		return nil
	}
	matches := agentsContextFilesRe.FindSubmatch(content)
	if len(matches) < 2 {
		return nil
	}
	return strings.Fields(strings.TrimSpace(string(matches[1])))
}

// getSourceFilesDiff returns a git diff scoped to source files only.
//
// If includePatterns is non-empty (read from AGENTS.md "## Context files"), they are
// passed as include pathspecs: `git diff HEAD -- <patterns>`.
// Otherwise, defaultContextExcludes are used to strip noise (config, docs, lock files)
// while keeping all source languages: `git diff HEAD -- :!*.yaml :!*.md …`.
//
// A 32 KB cap prevents runaway context sizes. When truncated, the most recent hunks
// are kept (most likely to contain the task-relevant additions) and a header notes it.
func getSourceFilesDiff(workDir string, includePatterns []string) string {
	args := []string{"diff", "HEAD", "--"}
	if len(includePatterns) > 0 {
		args = append(args, includePatterns...)
	} else {
		args = append(args, defaultContextExcludes...)
	}

	var out bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}

	result := out.String()
	const maxBytes = 32 * 1024
	if len(result) <= maxBytes {
		return result
	}

	// Keep the tail of the diff and align to the next file header so we never
	// emit a partial hunk.
	tail := result[len(result)-maxBytes:]
	if idx := strings.Index(tail, "\ndiff --git"); idx >= 0 {
		tail = tail[idx+1:]
	}
	return fmt.Sprintf("[diff truncated: showing last %d of %d bytes — earliest changes omitted]\n\n%s",
		len(tail), len(result), tail)
}

// runIntegrationCheck invokes the LLM with integration_check.md to catch
// cross-cutting issues (template↔struct mismatches, dead code, broken tests)
// that per-task verification gates miss. Returns an error if the LLM reports
// any BLOCKERs (exits non-zero).
func runIntegrationCheck(logger *slog.Logger, aiSessionHome, storyDescription, workDir string, contextPattern []string) error {
	if os.Getenv("IN_TEST_MODE") == "true" {
		logger.Info("Skipping integration check (test mode)")
		return nil
	}

	promptPath := filepath.Join(aiSessionHome, "headless", "session", "integration_check.md")
	promptTemplate, err := os.ReadFile(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Prompt file is optional; skip gracefully rather than failing the run.
			logger.Warn("integration_check.md not found — skipping integration check", "path", promptPath)
			return nil
		}
		return fmt.Errorf("reading integration_check.md: %w", err)
	}

	diff := getSourceFilesDiff(workDir, contextPattern)
	prompt := strings.ReplaceAll(string(promptTemplate), "{{story_description_here}}", storyDescription)
	prompt = strings.ReplaceAll(prompt, "{{codebase_diff_here}}", diff)

	if err := gemini.RunYolo(strings.NewReader(prompt), os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("integration check reported blockers (gemini exit: %w)", err)
	}
	return nil
}

// isRateLimitError reports whether the combined stdout+stderr output of a failed
// Gemini invocation indicates a transient rate-limit or quota error. These errors
// should be retried without consuming the task's attempt budget.
func isRateLimitError(output string) bool {
	lower := strings.ToLower(output)
	return strings.Contains(lower, "429") ||
		strings.Contains(lower, "rate limit") ||
		strings.Contains(lower, "ratelimitexceeded") ||
		strings.Contains(lower, "quota exceeded") ||
		strings.Contains(lower, "resource exhausted") ||
		strings.Contains(lower, "too many requests")
}

// executeTaskWithRetry invokes the LLM for a single task and retries on failure.
// In --yolo mode Gemini executes its own tool calls (run_shell_command, write_file, replace)
// live during the process — file changes happen inside the gemini call.
//
// A non-zero Gemini exit code is treated as a soft failure: verification still runs
// because file changes may have landed on disk. If verification passes, the task
// succeeds regardless of the Gemini exit code. If either fails, both errors are
// injected into the next attempt's prompt so the LLM can recover.
//
// retryDelay is applied before each retry to avoid LLM rate-limit errors that
// otherwise cause immediate failure on the following attempt.
func executeTaskWithRetry(logger *slog.Logger, featureDir, aiSessionHome, workDir, storyDescription, architectureDescription, sliceDescription, taskDescription, verificationCmd string, maxRetries int, retryDelay time.Duration, contextPattern []string) error {
	promptPath := filepath.Join(aiSessionHome, "headless", "session", "execute_task.md")

	// Read prompt template once; it does not change between retries.
	promptTemplate, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("reading execute_task.md: %w", err)
	}

	// attempt counts real failures against the budget. rateLimitRetries tracks
	// transient quota/rate-limit errors separately — they never consume the budget
	// but are capped to prevent an infinite loop.
	const maxRateLimitRetries = 20
	attempt := 1
	rateLimitRetries := 0
	var lastError error

	for attempt <= maxRetries {
		if (attempt > 1 || rateLimitRetries > 0) && retryDelay > 0 {
			logger.Info("Waiting before retry", "delay", retryDelay, "attempt", attempt, "rate_limit_retries", rateLimitRetries)
			time.Sleep(retryDelay)
		}

		// Capture the current diff before building the prompt so the LLM can see
		// all changes made by previous tasks (prevents field-name mismatches when
		// a later task references types introduced by an earlier one).
		changesSoFar := getSourceFilesDiff(workDir, contextPattern)

		promptContent := strings.ReplaceAll(string(promptTemplate), "{{story_description_here}}", storyDescription)
		promptContent = strings.ReplaceAll(promptContent, "{{architecture_description_here}}", architectureDescription)
		promptContent = strings.ReplaceAll(promptContent, "{{slice_description_here}}", sliceDescription)
		promptContent = strings.ReplaceAll(promptContent, "{{task_description_here}}", taskDescription)
		promptContent = strings.ReplaceAll(promptContent, "{{changes_so_far_here}}", changesSoFar)
		promptContent = strings.ReplaceAll(promptContent, "{{verification_command_here}}", verificationCmd)
		promptContent = strings.ReplaceAll(promptContent, "{{feature_dir_here}}", featureDir)
		if lastError != nil {
			promptContent = strings.ReplaceAll(promptContent, "{{error_message_here}}", lastError.Error())
			promptContent = strings.ReplaceAll(promptContent, "{{#if error_message}}", "")
			promptContent = strings.ReplaceAll(promptContent, "{{/if}}", "")
		} else {
			promptContent = errorBlockRe.ReplaceAllString(promptContent, "")
		}

		var geminiErr error
		var geminiOutput bytes.Buffer // captures stdout+stderr for rate-limit detection
		if os.Getenv("IN_TEST_MODE") != "true" {
			appendLog(logger, featureDir, fmt.Sprintf("Invoking LLM for task execution (attempt %d).", attempt))
			logger.Info("Invoking Gemini", "attempt", attempt)
			if err := gemini.RunYolo(strings.NewReader(promptContent), io.MultiWriter(os.Stdout, &geminiOutput), io.MultiWriter(os.Stderr, &geminiOutput)); err != nil {
				geminiErr = err
				appendLog(logger, featureDir, fmt.Sprintf("Gemini exited with error (attempt %d): %v — running verification anyway", attempt, err))
				logger.Warn("Gemini exited with error; running verification", "attempt", attempt, "error", err)
			}
		} else {
			logger.Info("Skipping Gemini invocation (test mode)", "attempt", attempt)
		}

		// Always run verification — file changes may have landed on disk even when
		// Gemini exited non-zero (e.g. due to a rate-limit or transient API error).
		var verificationOutput bytes.Buffer
		verifyCmd := exec.Command("sh", "-c", verificationCmd)
		verifyCmd.Stdout = &verificationOutput
		verifyCmd.Stderr = &verificationOutput
		verifyErr := verifyCmd.Run()

		if verifyErr == nil {
			appendLog(logger, featureDir, fmt.Sprintf("Verification passed (attempt %d).", attempt))
			logger.Info("Verification passed", "attempt", attempt)
			return nil
		}

		// Rate-limit / quota errors do not count against the attempt budget.
		if geminiErr != nil && isRateLimitError(geminiOutput.String()) {
			rateLimitRetries++
			msg := fmt.Sprintf("Rate-limit error on attempt %d (rate-limit retry %d/%d) — not counting against budget.", attempt, rateLimitRetries, maxRateLimitRetries)
			appendLog(logger, featureDir, msg)
			logger.Warn(msg)
			if rateLimitRetries >= maxRateLimitRetries {
				return fmt.Errorf("exceeded rate-limit retry budget (%d retries): %w", maxRateLimitRetries, geminiErr)
			}
			continue // attempt unchanged
		}

		// Real failure: build combined error for the next prompt and consume one attempt.
		var errParts []string
		if geminiErr != nil {
			errParts = append(errParts, geminiErr.Error())
		}
		errParts = append(errParts, verificationOutput.String())
		lastError = fmt.Errorf("%s", strings.Join(errParts, "\n"))

		appendLog(logger, featureDir, fmt.Sprintf("Verification failed (attempt %d): %s", attempt, verificationOutput.String()))
		logger.Error("Verification failed", "attempt", attempt)

		if attempt == maxRetries {
			return lastError
		}
		attempt++
	}

	return lastError
}

// depsMet reports whether all slices in deps are marked "done" in the plan.
func depsMet(p plan.Plan, deps []string) bool {
	for _, depID := range deps {
		found := false
		for _, s := range p {
			if s.ID == depID {
				found = true
				if s.Status != "done" {
					return false
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// extractVerificationCommand reads AGENTS.md from dir and extracts the verification
// command from a "## Verification\nRun: <cmd>" section.
func extractVerificationCommand(dir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		return "", fmt.Errorf("reading AGENTS.md: %w", err)
	}
	matches := agentsVerifyRe.FindSubmatch(content)
	if len(matches) < 2 {
		return "", fmt.Errorf("verification command not found in AGENTS.md (expected '## Verification\\nRun: <command>')")
	}
	return strings.TrimSpace(string(matches[1])), nil
}

// runShell executes a shell command, streaming output to stdout/stderr.
func runShell(shellCmd string) error {
	cmd := exec.Command("sh", "-c", shellCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
