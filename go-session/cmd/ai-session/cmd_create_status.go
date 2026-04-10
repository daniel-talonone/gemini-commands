package main

import (
	"fmt"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/status"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createStatusCmd)
	createStatusCmd.Flags().String("story-url", "", "URL of the story (e.g. Shortcut or Notion link)")
	createStatusCmd.Flags().String("mode", "", "Session mode")
	createStatusCmd.Flags().String("branch", "", "Git branch (defaults to current branch)")
	createStatusCmd.Flags().String("repo", "", "Repository slug org/repo (defaults to git remote origin)")
	createStatusCmd.Flags().String("work-dir", "", "Project working directory (defaults to git work-tree root)")
}

var createStatusCmd = &cobra.Command{
	Use:   "create-status <story-id>",
	Short: "Create a status.yaml file for a feature",
	Long: `Creates status.yaml inside the resolved feature directory for the given story ID.

Arguments:
  <story-id>  Story identifier (e.g. sc-1234) or an explicit path

The feature directory is resolved using the same logic as resolve-feature-dir.
Idempotent: if status.yaml already exists the command succeeds without modifying it.

Fields derived automatically when flags are omitted:
  --repo      from git remote origin
  --work-dir  from git work-tree root
  --branch    from current git branch

Errors:
  - Exactly 1 argument required
  - Feature directory cannot be resolved`,
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <story-id>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		featureDir, err := feature.ResolveFeatureDir(args[0], cwd, git.RemoteURL())
		if err != nil {
			return fmt.Errorf("resolving feature directory: %w", err)
		}

		storyURL, _ := cmd.Flags().GetString("story-url")
		mode, _ := cmd.Flags().GetString("mode")
		branch, _ := cmd.Flags().GetString("branch")
		repo, _ := cmd.Flags().GetString("repo")
		workDir, _ := cmd.Flags().GetString("work-dir")

		if repo == "" {
			repo = git.OrgRepo()
		}
		if workDir == "" {
			workDir = git.WorkDir()
		}
		if branch == "" {
			branch = git.CurrentBranch()
		}

		return status.Create(featureDir, repo, branch, workDir, storyURL, mode)
	},
}
