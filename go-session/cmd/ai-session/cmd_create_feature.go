package main

import (
	"fmt"
	"os"


	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	git "github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createFeatureCmd)
	createFeatureCmd.Flags().String("repo", "", "org/repo slug to write into status.yaml (derived from git remote if omitted)")
	createFeatureCmd.Flags().String("branch", "", "branch name to write into status.yaml (derived from git if omitted)")
	createFeatureCmd.Flags().String("work-dir", "", "repo root path for status.yaml (derived from git if omitted)")
}

var createFeatureCmd = &cobra.Command{
	Use:   "create-feature <feature-dir>",
	Short: "Create a feature directory with placeholder files",
	Long: `Creates a feature directory with the following placeholder files:
  plan.yml, questions.yml, review.yml — each containing []
  log.md  — # Work Log header
  pr.md   — # Pull Request header
  status.yaml — zero-value scaffold (repo/branch populated from git if available)

Arguments:
  <feature-dir>  Full path to the feature directory to create

Flags:
  --repo    org/repo slug (derived from git remote origin if omitted)
  --branch  branch name (derived from git if omitted)

Idempotent: exits 0 if the directory already exists (skips existing files).

Errors:
  - Exactly 1 argument required
  - Directory cannot be created (permissions, invalid path)`,
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

		if repo == "" {
			repo = git.OrgRepo()
		}
		if branch == "" {
			branch = git.CurrentBranch()
		}
		if workDir == "" {
			workDir = git.WorkDir()
		}

		if err := commands.CreateFeature(args[0], repo, branch, workDir); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("Created:", args[0])
		return nil
	},
}


