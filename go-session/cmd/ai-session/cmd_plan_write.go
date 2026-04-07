package main

import (
	"fmt"
	"io"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(planWriteCmd) }

var planWriteCmd = &cobra.Command{
	Use:   "plan-write <feature-dir>",
	Short: "Validate and write plan.yml from stdin",
	Long: `Reads a full plan YAML from stdin, validates it against the plan schema,
and writes it atomically to plan.yml in the feature directory.
The original bytes are preserved — no reformatting.

Arguments:
  <feature-dir>  Path to the feature directory

Schema requirements:
  - Top-level is a non-empty YAML sequence
  - Each slice: id (kebab-case, unique), description, status, tasks (non-empty)
  - Each task: id (kebab-case, unique across entire file), task body, status
  - Valid status values: todo, in-progress, done

Usage example:
  cat my-plan.yml | ai-session plan-write /path/to/feature-dir
  printf '%s' "$PLAN_YAML" | ai-session plan-write /path/to/feature-dir

Errors:
  - Invalid YAML
  - Schema violations (missing fields, bad status, non-kebab-case ids, duplicates)
  - Exactly 1 positional argument required`,
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <feature-dir>, got %d", len(args))
		}
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading stdin:", err)
			os.Exit(1)
		}
		if err := plan.WritePlan(args[0], data); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("plan.yml written successfully.")
		return nil
	},
}
