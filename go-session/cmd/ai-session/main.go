package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-session",
	Short: "Local file operations for the ai-session workflow",
	Long: `ai-session manages feature directories for the ai-session workflow.

It provides deterministic, testable replacements for shell scripts and yq
one-liners. Designed to be invoked by LLMs — every subcommand has strict
input validation and --help output sufficient to use without reading source.

Feature directory: a directory containing description.md, plan.yml,
questions.yml, review.yml, log.md, and pr.md for a single story.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
