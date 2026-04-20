package main

import (
	"fmt"
	"io"
	"os"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var splitSlice string
var splitTask string

func init() {
	planSplitTaskCmd.Flags().StringVar(&splitSlice, "slice", "", "Slice ID containing the task (required)")
	planSplitTaskCmd.Flags().StringVar(&splitTask, "task", "", "Task ID to split (required)")
	_ = planSplitTaskCmd.MarkFlagRequired("slice")
	_ = planSplitTaskCmd.MarkFlagRequired("task")
	rootCmd.AddCommand(planSplitTaskCmd)
}

var planSplitTaskCmd = &cobra.Command{
	Use:   "plan-split-task <feature-dir> --slice <id> --task <id>",
	Short: "Replace a single todo task with N atomic subtasks",
	Long: `Reads a YAML list of {suffix, task} objects from stdin and replaces the
specified task with N new tasks. Generated IDs are "{original-id}-{suffix}".
All replacement tasks get status: todo.

The original task must have status todo — done and in-progress tasks are protected.
At least 2 replacement tasks are required.

Arguments:
  <feature-dir>  Path to the feature directory containing plan.yml

Flags:
  --slice  Slice ID containing the task (required)
  --task   Task ID to split (required)

stdin: YAML list of replacement tasks, e.g.:
  - suffix: readme
    task: Update README.md Dependencies section
  - suffix: agents
    task: Update AGENTS.md State Management section

Errors:
  - Fewer than 2 replacements
  - Non-kebab-case suffix
  - Task body contains id: or status: lines (injection guard)
  - Generated ID collides with an existing task ID
  - Slice or task not found
  - Task status is not todo`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <feature-dir>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading stdin:", err)
			os.Exit(1)
		}
		var replacements []commands.SplitTaskEntry
		if err := yaml.Unmarshal(data, &replacements); err != nil {
			fmt.Fprintln(os.Stderr, "Error: stdin must be a valid YAML list of {suffix, task} objects:", err)
			os.Exit(1)
		}
		if err := commands.SplitTask(args[0], splitSlice, splitTask, replacements); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Task %q in slice %q split into %d tasks.\n", splitTask, splitSlice, len(replacements))
		return nil
	},
}
