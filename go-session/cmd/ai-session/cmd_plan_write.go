package main

import (
	"fmt"
	"io"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/spf13/cobra"
)

var architectureFlag bool

var planWriteCmd = &cobra.Command{
	Use:   "write [feature-id]",
	Short: "Validate and write plan.yml or architecture.md from stdin",
	Long: `Reads a full plan YAML or architecture markdown from stdin, validates it,
and writes it atomically to plan.yml or architecture.md in the feature directory.
The original bytes are preserved — no reformatting.

Arguments:
  [feature-id]  The feature ID (e.g., sc-12345, notion-xxxx)

Flags:
  --architecture  Write to architecture.md instead of plan.yml

Schema requirements for plan.yml:
  - Top-level is a non-empty YAML sequence
  - Each slice: id (kebab-case, unique), description, status, tasks (non-empty)
  - Each task: id (kebab-case, unique across entire file), task body, status
  - Valid status values: todo, in-progress, done

Usage examples:
  cat my-plan.yml | ai-session plan write sc-12345
  printf '%s' "$ARCH_MD" | ai-session plan write --architecture sc-12345

Errors:
  - Invalid YAML (for plans)
  - Schema violations (missing fields, bad status, non-kebab-case ids, duplicates)
  - Exactly 1 positional argument required`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		featureID := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current working directory: %w", err)
		}
		remoteURL := git.RemoteURL()

		featureDir, err := feature.ResolveFeatureDir(featureID, cwd, remoteURL)
		if err != nil {
			return fmt.Errorf("resolving feature dir: %w", err)
		}

		if _, err := os.Stat(featureDir); os.IsNotExist(err) {
			return fmt.Errorf("feature directory not found: %s", featureDir)
		}

		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading from stdin: %w", err)
		}

		if architectureFlag {
			if err := plan.WriteArchitecture(featureDir, string(data)); err != nil {
				return fmt.Errorf("writing architecture: %w", err)
			}
			fmt.Println("architecture.md written successfully.")
		} else {
			if err := plan.WritePlan(featureDir, data); err != nil {
				return fmt.Errorf("writing plan: %w", err)
			}
			fmt.Println("plan.yml written successfully.")
		}

		return nil
	},
}

func init() {
	planCmd.AddCommand(planWriteCmd)
	planWriteCmd.Flags().BoolVar(&architectureFlag, "architecture", false, "Write to architecture.md instead of plan.yml")
}
