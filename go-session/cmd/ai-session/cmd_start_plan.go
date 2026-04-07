package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/daniel-talonone/gemini-commands/internal/commands/status"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startPlanCmd)
}

const enrichScript = "scripts/enrich_tasks.sh"

var startPlanCmd = &cobra.Command{
	Use:   "start-plan <story-id>",
	Short: "Start the headless plan for a feature",
	Long: `Executes the headless plan for a given feature.
This command replaces the 'orchestrate.sh --plan' script.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		storyID := args[0]

		logger.Info("Resolving feature directory", "story_id", storyID)
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		remoteURL := git.RemoteURL()
		featureDir, err := commands.ResolveFeatureDir(storyID, cwd, remoteURL)
		if err != nil {
			return err
		}
		logger.Info("Feature directory resolved", "path", featureDir)

		logger.Info("Verifying preconditions")
		descPath := filepath.Join(featureDir, "description.md")
		if _, err := os.Stat(descPath); os.IsNotExist(err) {
			return fmt.Errorf("description.md not found in %s — run /session:new or /session:define first", featureDir)
		}
		logger.Info("Preconditions verified")

		logger.Info("Updating status to 'plan'")
		repo := git.OrgRepo()
		branch := git.CurrentBranch()
		if err := status.Write(featureDir, "plan", repo, branch); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}
		logger.Info("Status updated successfully")

		logger.Info("Preparing plan prompt")
		prompt, err := preparePlanPrompt(storyID)
		if err != nil {
			return fmt.Errorf("preparing plan prompt: %w", err)
		}
		logger.Info("Plan prompt prepared")

		logger.Info("Executing gemini command")
		geminiCmd := exec.Command("gemini", "--yolo", "-p", prompt)
		geminiCmd.Stdout = os.Stdout
		geminiCmd.Stderr = os.Stderr

		if err := geminiCmd.Run(); err != nil {
			logger.Error("Gemini command failed", "error", err)
			if writeErr := status.Write(featureDir, "plan-failed", repo, branch); writeErr != nil {
				logger.Error("Failed to write plan-failed status", "error", writeErr)
			}
			return fmt.Errorf("plan generation failed")
		}

		logger.Info("Gemini command finished successfully")
		if err := status.Write(featureDir, "plan-done", repo, branch); err != nil {
			return fmt.Errorf("updating status to plan-done: %w", err)
		}
		logger.Info("Status updated to plan-done")
		fmt.Println("Plan generation successful.")

		// Trigger enrichment synchronously
		logger.Info("Triggering synchronous enrichment")
		enrichScriptPath := filepath.Join(getAISessionHome(), enrichScript)
		enrichCmd := exec.Command(enrichScriptPath, featureDir)
		enrichCmd.Stdout = os.Stdout
		enrichCmd.Stderr = os.Stderr
		if err := enrichCmd.Run(); err != nil {
			logger.Error("Enrichment script failed", "error", err)
			return fmt.Errorf("enrichment failed: %w", err)
		}
		logger.Info("Synchronous enrichment finished successfully")

		return nil
	},
}

func preparePlanPrompt(storyID string) (string, error) {
	aiSessionHome := getAISessionHome()

	promptPath := filepath.Join(aiSessionHome, "headless", "session", "plan.md")
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("reading prompt file at %s: %w", promptPath, err)
	}

	prompt := strings.ReplaceAll(string(content), "{{args}}", storyID)
	return prompt, nil
}

func getAISessionHome() string {
	aiSessionHome := os.Getenv("AI_SESSION_HOME")
	if aiSessionHome == "" {
		executable, err := os.Executable()
		if err != nil {
			return "" // Should not happen in normal operation
		}
		// Assumes the executable is in go-session/bin/
		aiSessionHome = filepath.Join(filepath.Dir(executable), "..", "..")
	}
	return aiSessionHome
}
