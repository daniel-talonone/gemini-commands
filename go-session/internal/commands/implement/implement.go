package implement

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/daniel-talonone/gemini-commands/internal/commands/plan"
	"github.com/daniel-talonone/gemini-commands/internal/commands/status"
)

// errorBlockRe strips the entire {{#if error_message}}...{{/if}} block (including
// its XML wrapper) when there is no error to inject.
var errorBlockRe = regexp.MustCompile(`(?s)\{\{#if error_message\}\}.*?\{\{/if\}\}`)

// agentsVerifyRe extracts the verification command from an AGENTS.md "## Verification" section.
var agentsVerifyRe = regexp.MustCompile(`(?m)^## Verification\r?\nRun: (.+)$`)

// appendLog writes a timestamped entry to log.md. Logging is best-effort — a
// failure is reported via the logger but never stops the orchestration.
func appendLog(logger *slog.Logger, featureDir, msg string) {
	if err := commands.AppendLog(featureDir, msg); err != nil {
		logger.Warn("failed to append log entry", "error", err)
	}
}

// Run orchestrates the full implementation loop for a feature: reads the plan,
// runs the verification gate, iterates slices/tasks, invokes the LLM per task,
// and marks statuses as it progresses.
//
// workDir must be the target project root (the directory containing AGENTS.md).
// aiSessionHome must be the resolved AI_SESSION_HOME path.
func Run(logger *slog.Logger, featureID, featureDir, workDir, aiSessionHome string) error {
	appendLog(logger, featureDir, fmt.Sprintf("--- Starting implementation orchestration for feature: %s ---", featureID))
	logger.Info("Starting implementation orchestration", "feature_id", featureID)

	// Extract verification command from the project's AGENTS.md.
	verificationCmd, err := extractVerificationCommand(workDir)
	if err != nil {
		return fmt.Errorf("extracting verification command: %w", err)
	}
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
	storyDescBytes, err := os.ReadFile(filepath.Join(featureDir, "description.md"))
	if err != nil {
		return fmt.Errorf("reading description.md: %w", err)
	}
	storyDescription := string(storyDescBytes)

	// Load architecture if present — gives the LLM design constraints and pattern refs.
	architectureDescription, err := plan.LoadArchitecture(featureDir)
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

			// Process each task in the slice sequentially.
			for j := range s.Tasks {
				t := &s.Tasks[j]
				if t.Status == "done" {
					continue
				}

				resuming := t.Status == "in-progress"
				if resuming {
					appendLog(logger, featureDir, fmt.Sprintf("Resuming task: %s (was in-progress from a prior run)", t.ID))
				} else {
					appendLog(logger, featureDir, fmt.Sprintf("Starting task: %s", t.ID))
				}
				logger.Info("Starting task", "slice", s.ID, "task", t.ID, "resuming", resuming)

				if err := plan.UpdateTask(featureDir, t.ID, "in-progress"); err != nil {
					return fmt.Errorf("updating task %s to in-progress: %w", t.ID, err)
				}

				if err := executeTaskWithRetry(logger, featureDir, aiSessionHome, storyDescription, architectureDescription, s.Description, t.Task, verificationCmd); err != nil {
					appendLog(logger, featureDir, fmt.Sprintf("Task %s FAILED after all retries: %v", t.ID, err))
					logger.Error("Task failed after all retries", "task", t.ID, "error", err)
					return fmt.Errorf("task %s in slice %s failed: %w", t.ID, s.ID, err)
				}

				if err := plan.UpdateTask(featureDir, t.ID, "done"); err != nil {
					return fmt.Errorf("updating task %s to done: %w", t.ID, err)
				}
				t.Status = "done"
				appendLog(logger, featureDir, fmt.Sprintf("Task %s completed successfully.", t.ID))
				logger.Info("Task completed", "task", t.ID)
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

	if err := status.Write(featureDir, "implement-done", "", ""); err != nil {
		return fmt.Errorf("updating status to implement-done: %w", err)
	}
	appendLog(logger, featureDir, "--- IMPLEMENT COMPLETE ---")
	logger.Info("IMPLEMENT COMPLETE")
	return nil
}

// executeTaskWithRetry invokes the LLM for a single task and retries on verification failure.
// In --yolo mode Gemini executes its own tool calls (run_shell_command, write_file, replace)
// live during the process — file changes happen inside the gemini call. We stream its output
// to the terminal and check the exit code; there is nothing to capture or re-execute.
func executeTaskWithRetry(logger *slog.Logger, featureDir, aiSessionHome, storyDescription, architectureDescription, sliceDescription, taskDescription, verificationCmd string) error {
	const maxRetries = 5
	promptPath := filepath.Join(aiSessionHome, "headless", "session", "execute_task.md")

	// Read prompt template once; it does not change between retries.
	promptTemplate, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("reading execute_task.md: %w", err)
	}

	var lastVerificationError error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		promptContent := strings.ReplaceAll(string(promptTemplate), "{{story_description_here}}", storyDescription)
		promptContent = strings.ReplaceAll(promptContent, "{{architecture_description_here}}", architectureDescription)
		promptContent = strings.ReplaceAll(promptContent, "{{slice_description_here}}", sliceDescription)
		promptContent = strings.ReplaceAll(promptContent, "{{task_description_here}}", taskDescription)

		if lastVerificationError != nil {
			// Inject error and strip the conditional markers.
			promptContent = strings.ReplaceAll(promptContent, "{{error_message_here}}", lastVerificationError.Error())
			promptContent = strings.ReplaceAll(promptContent, "{{#if error_message}}", "")
			promptContent = strings.ReplaceAll(promptContent, "{{/if}}", "")
		} else {
			// Strip the entire conditional block (including surrounding XML tags).
			promptContent = errorBlockRe.ReplaceAllString(promptContent, "")
		}

		if os.Getenv("IN_TEST_MODE") != "true" {
			appendLog(logger, featureDir, fmt.Sprintf("Invoking LLM for task execution (attempt %d).", attempt))
			logger.Info("Invoking Gemini", "attempt", attempt)
			// Pass the prompt via stdin to avoid OS argument-length limits.
			// In --yolo mode Gemini executes tool calls autonomously; changes land on disk
			// during this call. Stream output to the terminal and check exit code only.
			geminiCmd := exec.Command("gemini", "--yolo")
			geminiCmd.Stdin = strings.NewReader(promptContent)
			geminiCmd.Stdout = os.Stdout
			geminiCmd.Stderr = os.Stderr
			if err := geminiCmd.Run(); err != nil {
				return fmt.Errorf("gemini failed on attempt %d: %w", attempt, err)
			}
		} else {
			logger.Info("Skipping Gemini invocation (test mode)", "attempt", attempt)
		}

		var verificationOutput bytes.Buffer
		cmd := exec.Command("sh", "-c", verificationCmd)
		cmd.Stdout = &verificationOutput
		cmd.Stderr = &verificationOutput
		if err := cmd.Run(); err != nil {
			lastVerificationError = fmt.Errorf("attempt %d failed:\n%s", attempt, verificationOutput.String())
			appendLog(logger, featureDir, fmt.Sprintf("Verification failed (attempt %d): %s", attempt, verificationOutput.String()))
			logger.Error("Verification failed", "attempt", attempt)
			if attempt == maxRetries {
				return lastVerificationError
			}
			continue
		}

		appendLog(logger, featureDir, fmt.Sprintf("Verification passed (attempt %d).", attempt))
		logger.Info("Verification passed", "attempt", attempt)
		return nil
	}

	return lastVerificationError
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
