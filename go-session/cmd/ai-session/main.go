package main

import (
	"fmt"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/llm"
	"github.com/spf13/cobra"
)

var modelFlag string

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

func init() {
	rootCmd.PersistentFlags().StringVar(&modelFlag, "model", "gemini", "LLM backend: gemini, gemini-flash, or claude")
}

// getRunner returns the Runner selected by the --model flag.
func getRunner() (llm.Runner, error) {
	return llm.NewRunner(llm.Model(modelFlag))
}

var OsExit = os.Exit

func ExecuteCobra() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func main() {
	OsExit(ExecuteCobra())
}

// GetRootCmd returns the root cobra command for testing purposes.
func GetRootCmd() *cobra.Command {
	return rootCmd
}
