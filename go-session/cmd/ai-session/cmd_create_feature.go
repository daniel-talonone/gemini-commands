package main

import (
	"fmt"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createFeatureCmd)
	createFeatureCmd.Flags().String("repo", "", "Repository slug org/repo (overrides git remote detection)")
	createFeatureCmd.Flags().String("branch", "", "Git branch (overrides git detection)")
	createFeatureCmd.Flags().String("work-dir", "", "Project working directory (overrides git detection)")
}

var createFeatureCmd = &cobra.Command{
	Use:   "create-feature <feature-dir>",
	Short: "Scaffold a feature directory with placeholder files",
	Long: `Creates a feature directory and populates it with placeholder files.

Arguments:
  <feature-dir>  Full path to the feature directory (created if absent)

Files created (existing files are never overwritten):
  plan.yml       empty plan
  questions.yml  empty questions list
  review.yml     empty review findings list
  log.md         work log with header
  pr.md          pull request placeholder
  status.yaml    feature status (repo/branch/work_dir derived from git)

Idempotent: safe to run on an existing directory.

Errors:
  - Exactly 1 argument required`,
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <feature-dir>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		branch, _ := cmd.Flags().GetString("branch")
		workDir, _ := cmd.Flags().GetString("work-dir")
		return feature.CreateFeature(args[0], repo, branch, workDir)
	},
}
