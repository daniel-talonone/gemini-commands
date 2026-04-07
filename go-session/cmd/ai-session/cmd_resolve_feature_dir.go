package main

import (
	"fmt"
	"os"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	git "github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(resolveFeatureDirCmd)
}

var resolveFeatureDirCmd = &cobra.Command{
	Use:   "resolve-feature-dir <story-id>",
	Short: "Resolve the full path to a feature directory",
	Long: `Resolves the feature directory path for the given story ID.

Arguments:
  <story-id>  Story identifier (e.g. sc-1234) or an explicit path

Resolution order:
  1. story-id contains "/" or starts with "." or "~": returned as-is
  2. .features/<story-id>/ exists in CWD: return that path (legacy layout)
  3. Derive from git remote origin: ~/.ai-session/features/<org>/<repo>/<story-id>

Prints the resolved path to stdout (no trailing newline).

Errors:
  - Not in a git repository and no local .features/<story-id> found
  - Git remote URL cannot be parsed into org/repo
  - Exactly 1 argument required`,
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <story-id>, got %d", len(args))
		}
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting working directory:", err)
			os.Exit(1)
		}
		result, err := commands.ResolveFeatureDir(args[0], cwd, git.RemoteURL())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Print(result)
		return nil
	},
}
