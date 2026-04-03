package main

import (
	"fmt"
	"os"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appendLogCmd)
}

var appendLogCmd = &cobra.Command{
	Use:   "append-log <feature-dir> <message>",
	Short: "Append a timestamped entry to log.md",
	Long: `Appends a timestamped Markdown entry to log.md in the feature directory.

Arguments:
  <feature-dir>  Path to the feature directory (must exist)
  <message>      Log message text (quote multi-word messages)

Entry format written to log.md:
  ## [2026-01-01T10:00:00Z]

  <message>

Entries are separated by a blank line. Creates log.md if it does not exist.

Errors:
  - Feature directory does not exist
  - Exactly 2 arguments required`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("exactly 2 arguments required: <feature-dir> and <message>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := commands.AppendLog(args[0], args[1]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return nil
	},
}
