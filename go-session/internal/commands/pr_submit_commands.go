package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/github"
	"github.com/daniel-talonone/gemini-commands/internal/llm" // New import
	"github.com/daniel-talonone/gemini-commands/internal/pr"
	"github.com/daniel-talonone/gemini-commands/internal/status"
	"github.com/spf13/cobra"
)

// ResolveFeatureDir is a variable to allow mocking in tests. Exported for testing.
var ResolveFeatureDir = feature.ResolveFeatureDir

// GetRunner is a package-level variable for the LLM runner, allowing it to be mocked or set by the main package.
// It defaults to an implementation that returns an error if not set.
var GetRunner = func() (llm.Runner, error) {
	return nil, fmt.Errorf("LLM runner not set for commands package")
}

// PrSubmitCmd represents the submit command. Exported for main package to use.
var PrSubmitCmd = &cobra.Command{
	Use:   "submit <story-id>",
	Short: "Submits a GitHub PR and updates feature state",
	Long: `Submits a GitHub pull request using the PR description from pr.md.

The command:
1. Resolves the feature directory for the given story-id
2. Validates that pr.md is present and non-empty
3. Checks that no PR has already been submitted (pr_url in status.yaml)
4. Uses the LLM to generate a conventional-commit-style title from pr.md content
5. Calls github.CreatePR with the generated title
6. Updates status.yaml with the PR URL and sets pipeline_step to pr-submitted

Fails with a clear error if:
- pr.md is missing or empty
- A PR has been submitted (pr_url already in status.yaml)
- Title generation fails
- PR creation fails`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		storyId := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current working directory: %w", err)
		}

		featureDir, err := ResolveFeatureDir(storyId, cwd, git.RemoteURL())
		if err != nil {
			return fmt.Errorf("resolving feature directory: %w", err)
		}

		// Load status to check if PR already exists
		s, err := status.LoadStatus(featureDir)
		if err != nil {
			return fmt.Errorf("loading status.yaml: %w", err)
		}

		// Check if PR was already submitted
		if s.PRURL != "" {
			return fmt.Errorf("PR already submitted for story %s: %s", storyId, s.PRURL)
		}

		// Read pr.md and validate it's non-empty
		prContent, err := pr.Read(featureDir)
		if err != nil {
			return fmt.Errorf("reading pr.md: %w", err)
		}

		if strings.TrimSpace(prContent) == "" {
			return fmt.Errorf("pr.md is missing or empty for story %s — run 'ai-session create-pr-description %s' first", storyId, storyId)
		}

		if strings.TrimSpace(prContent) == "# Pull Request" {
			return fmt.Errorf("pr.md contains default content. Please provide a description for story %s", storyId)
		}

		// Generate title using LLM
		title, err := GeneratePRTitle(prContent, getAISessionHome()) // Pass getAISessionHome()
		if err != nil {
			return fmt.Errorf("generating PR title: %w", err)
		}

		// Create PR using existing functions
		base := git.DefaultBranch()
		head := git.CurrentBranch()
		if head == "" {
			return fmt.Errorf("unable to determine current git branch")
		}

		prURL, err := github.CreatePR(s.WorkDir, base, head, title, prContent)
		if err != nil {
			return fmt.Errorf("creating PR: %w", err)
		}

		// Update status with PR URL
		if err := status.WritePRURL(featureDir, prURL); err != nil {
			return fmt.Errorf("updating status.yaml: %w", err)
		}

		fmt.Printf("Pull request submitted successfully: %s\n", prURL)
		return nil
	},
}

// GeneratePRTitle uses the LLM to extract a conventional-commit-style title from pr.md content.
// Exported for testing purposes.
var GeneratePRTitle = func(prContent string, aiSessionHome string) (string, error) {
	promptTemplatePath := filepath.Join(aiSessionHome, "headless", "session", "pr-submit-title.md")
	promptTemplate, err := os.ReadFile(promptTemplatePath)
	if err != nil {
		return "", fmt.Errorf("reading prompt template: %w", err)
	}

	promptContent := strings.ReplaceAll(string(promptTemplate), "{{pr_content}}", prContent)

	runner, err := GetRunner()
	if err != nil {
		return "", fmt.Errorf("invalid --model flag: %w", err)
	}

	// Capture LLM output into a buffer
	var out bytes.Buffer
	var stderr bytes.Buffer
	if err := runner.Run(strings.NewReader(promptContent), &out, &stderr); err != nil {
		return "", fmt.Errorf("LLM error: %w", err)
	}

	// Trim whitespace and newlines
	title := strings.TrimSpace(out.String())
	if title == "" {
		return "", fmt.Errorf("LLM generated empty title")
	}

	return title, nil
}

// getAISessionHome needs to be implemented or passed from main package.
// For now, it will be a placeholder.
func getAISessionHome() string {
	// This will eventually be passed from the main package or retrieved from an environment variable.
	return os.Getenv("AI_SESSION_HOME")
}

func init() {
	// prCmd is part of the main package, so we cannot AddCommand here.
	// This will be done in the main package where prCmd resides.
}