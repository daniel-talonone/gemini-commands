package main

import (
	"fmt"
	"os"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createFeatureCmd)
}

var createFeatureCmd = &cobra.Command{
	Use:   "create-feature <feature-dir>",
	Short: "Create a feature directory with placeholder files",
	Long: `Creates a feature directory with the following placeholder files:
  plan.yml, questions.yml, review.yml — each containing []
  log.md  — # Work Log header
  pr.md   — # Pull Request header

Arguments:
  <feature-dir>  Full path to the feature directory to create

Idempotent: exits 0 if the directory already exists (rewrites placeholder files).

Errors:
  - Exactly 1 argument required
  - Directory cannot be created (permissions, invalid path)`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <feature-dir>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := commands.CreateFeature(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("Created:", args[0])
		return nil
	},
}
