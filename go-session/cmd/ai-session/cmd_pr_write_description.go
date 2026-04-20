package main

import (
	"fmt"
	"io"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/pr"
	"github.com/daniel-talonone/gemini-commands/internal/status"
	"github.com/spf13/cobra"
)

var prWriteDescriptionCmd = &cobra.Command{
	Use:   "write-description <story-id> [<description>]",
	Short: "Write PR description from stdin or positional argument",
	Long: `Write a PR description to pr.md in the resolved feature directory.

Accepts the description either from stdin or as a positional argument (not both).
If both are provided simultaneously, exits with an error.
If neither are provided, exits with an error.

Examples:
  echo "PR description here" | ai-session pr write-description my-story-id
  ai-session pr write-description my-story-id "PR description here"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 2 {
			return fmt.Errorf("accepts 1 or 2 arguments, received %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		storyId := args[0]

		// Detect if stdin is piped (not a TTY)
		stdinStat, err := os.Stdin.Stat()
		if err != nil {
			return fmt.Errorf("checking stdin: %w", err)
		}
		stdinPiped := (stdinStat.Mode() & os.ModeCharDevice) == 0

		hasArg := len(args) == 2

		if stdinPiped && hasArg {
			return fmt.Errorf("ambiguous input: both stdin and positional argument provided")
		}

		// Determine description source
		var description string
		switch {
		case hasArg:
			description = args[1]
		case stdinPiped:
			stdinData, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			if len(stdinData) == 0 {
				return fmt.Errorf("neither stdin nor positional argument provided")
			}
			description = string(stdinData)
		default:
			return fmt.Errorf("neither stdin nor positional argument provided")
		}

		// Resolve feature directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current working directory: %w", err)
		}

		featureDir, err := feature.ResolveFeatureDir(storyId, cwd, git.RemoteURL())
		if err != nil {
			return fmt.Errorf("resolving feature directory for story %q: %w", storyId, err)
		}

		// Write PR description
		if err := pr.Write(featureDir, description); err != nil {
			return fmt.Errorf("writing PR description: %w", err)
		}

		// Update status
		if err := status.Write(featureDir, "pr-description-done", git.OrgRepo(), git.CurrentBranch()); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}

		fmt.Printf("PR description written successfully for story %s\n", storyId)
		return nil
	},
}

func init() {
	prCmd.AddCommand(prWriteDescriptionCmd)
}
