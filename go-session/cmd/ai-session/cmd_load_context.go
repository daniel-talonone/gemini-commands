package main

import (
	"fmt"
	"os"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loadContextCmd)
}

var loadContextCmd = &cobra.Command{
	Use:   "load-context <story-id>",
	Short: "Load all feature context files as XML blocks for LLM consumption",
	Long: `Resolves the feature directory for <story-id> and outputs all .md, .yml, .yaml
files (excluding _* files) as XML blocks:

  <file name="description.md">
  ...content...
  </file>

  <file name="plan.yml">
  ...content...
  </file>

Files are sorted alphabetically. Output is printed to stdout with no trailing newline.

Arguments:
  <story-id>  Story ID or explicit path (same resolution as resolve-feature-dir)

Resolution order:
  1. story-id contains "/" or starts with "." or "~": used as-is
  2. .features/<story-id>/ exists in CWD: use that path (legacy layout)
  3. Derive from git remote: ~/.ai-session/features/<org>/<repo>/<story-id>

Errors:
  - Feature directory does not exist after resolution
  - Exactly 1 argument required`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <story-id>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting working directory:", err)
			os.Exit(1)
		}
		featureDir, err := commands.ResolveFeatureDir(args[0], cwd, gitRemoteURL())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		output, err := commands.LoadContext(featureDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Print(output)
		return nil
	},
}
