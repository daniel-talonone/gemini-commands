package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/spf13/cobra"
)

var enrichSlice string
var enrichTask string

func init() {
	planEnrichTaskCmd.Flags().StringVar(&enrichSlice, "slice", "", "Slice ID containing the task (required)")
	planEnrichTaskCmd.Flags().StringVar(&enrichTask, "task", "", "Task ID to enrich (required)")
	_ = planEnrichTaskCmd.MarkFlagRequired("slice")
	_ = planEnrichTaskCmd.MarkFlagRequired("task")
	rootCmd.AddCommand(planEnrichTaskCmd)
}

var planEnrichTaskCmd = &cobra.Command{
	Use:   "plan-enrich-task <feature-dir> --slice <id> --task <id>",
	Short: "Update the task body of a single todo task in plan.yml",
	Long: `Reads a new task description from stdin and updates only the task: field
of the specified task using the yaml.Node API.
All other fields (id, status, other tasks, slices) are preserved exactly.

The task must have status todo — done and in-progress tasks are protected.

Arguments:
  <feature-dir>  Path to the feature directory containing plan.yml

Flags:
  --slice  Slice ID containing the task (required)
  --task   Task ID to update (required)

stdin: plain text task description.
  Must NOT contain lines starting with "id:" or "status:" — the injection
  guard rejects these to prevent LLMs from accidentally overwriting YAML fields.

Usage example:
  echo "FILE: src/foo.go — add the Foo function" | \
    ai-session plan-enrich-task /path/to/dir --slice my-slice --task my-task

Errors:
  - Empty body
  - Body contains id: or status: lines (injection guard)
  - Slice or task not found
  - Task status is not todo (done/in-progress tasks are protected)`,
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
		body := strings.TrimSpace(string(data))
		if err := commands.EnrichTask(args[0], enrichSlice, enrichTask, body); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Task %q in slice %q updated.\n", enrichTask, enrichSlice)
		return nil
	},
}
