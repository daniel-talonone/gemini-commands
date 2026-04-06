package main

import (
	"fmt"
	"os"

	plan "github.com/daniel-talonone/gemini-commands/internal/commands/plan"
	"github.com/spf13/cobra"
)

var updateTaskStatus string
var updateSliceStatus string

func init() {
	updateTaskCmd.Flags().StringVar(&updateTaskStatus, "status", "", "New status: todo, in-progress, or done (required)")
	_ = updateTaskCmd.MarkFlagRequired("status")

	updateSliceCmd.Flags().StringVar(&updateSliceStatus, "status", "", "New status: todo, in-progress, or done (required)")
	_ = updateSliceCmd.MarkFlagRequired("status")

	rootCmd.AddCommand(updateTaskCmd)
	rootCmd.AddCommand(updateSliceCmd)
}

var updateTaskCmd = &cobra.Command{
	Use:   "update-task <feature-dir> <task-id> --status <todo|in-progress|done>",
	Short: "Update the status of a task in plan.yml",
	Long: `Updates the status field of a task in plan.yml using the yaml.Node API.
All other YAML content (descriptions, task text, other statuses) is preserved exactly.

Arguments:
  <feature-dir>  Path to the feature directory containing plan.yml
  <task-id>      Kebab-case task ID as it appears in plan.yml (unique across all slices)

Flags:
  --status  New status value: todo, in-progress, or done (required)

Errors:
  - plan.yml not found in <feature-dir>
  - <task-id> not found in any slice
  - --status is not one of: todo, in-progress, done
  - Exactly 2 positional arguments required`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("exactly 2 arguments required: <feature-dir> and <task-id>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := plan.UpdateTask(args[0], args[1], updateTaskStatus); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return nil
	},
}

var updateSliceCmd = &cobra.Command{
	Use:   "update-slice <feature-dir> <slice-id> --status <todo|in-progress|done>",
	Short: "Update the status of a slice in plan.yml",
	Long: `Updates the status field of a top-level slice in plan.yml using the yaml.Node API.
All other YAML content is preserved exactly.

Arguments:
  <feature-dir>  Path to the feature directory containing plan.yml
  <slice-id>     Kebab-case slice ID as it appears in plan.yml

Flags:
  --status  New status value: todo, in-progress, or done (required)

Errors:
  - plan.yml not found in <feature-dir>
  - <slice-id> not found
  - --status is not one of: todo, in-progress, done
  - Exactly 2 positional arguments required`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("exactly 2 arguments required: <feature-dir> and <slice-id>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := plan.UpdateSlice(args[0], args[1], updateSliceStatus); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return nil
	},
}
