package main

import (
	"fmt"
	"os"
	"strings"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/spf13/cobra"
)

var planGetSlice string
var planGetTask string

func init() {
	planGetCmd.Flags().StringVar(&planGetSlice, "slice", "", "Slice ID to retrieve (required)")
	planGetCmd.Flags().StringVar(&planGetTask, "task", "", "Task ID to retrieve (optional; returns full task body)")
	_ = planGetCmd.MarkFlagRequired("slice")
	rootCmd.AddCommand(planGetCmd)
}

var planGetCmd = &cobra.Command{
	Use:   "plan-get <feature-dir> --slice <slice-id> [--task <task-id>]",
	Short: "Get full details of a slice or a single task from plan.yml",
	Long: `Returns the full details of one slice or one task from plan.yml.

Without --task: prints the slice header (id, description, status, depends_on)
  and a compact task list (id + status only, no bodies).
With --task <id>: prints the complete body of that one task.

Arguments:
  <feature-dir>  Path to the feature directory containing plan.yml

Flags:
  --slice  Slice ID (required)
  --task   Task ID within the slice (optional)

Errors:
  - plan.yml not found in <feature-dir>
  - Slice not found
  - Task not found within the slice
  - Exactly 1 positional argument required`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <feature-dir>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		featureDir := args[0]

		if planGetTask == "" {
			s, err := commands.GetSlice(featureDir, planGetSlice)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Printf("id:          %s\n", s.ID)
			fmt.Printf("status:      %s\n", s.Status)
			fmt.Printf("description: %s\n", s.Description)
			if len(s.DependsOn) > 0 {
				fmt.Printf("depends_on:  %s\n", strings.Join(s.DependsOn, ", "))
			}
			fmt.Println()
			fmt.Println("tasks:")
			tasks, err := commands.ListTasks(featureDir, planGetSlice)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			for _, t := range tasks {
				fmt.Printf("  %-38s [%s]\n", t.ID, t.Status)
			}
		} else {
			t, err := commands.GetTask(featureDir, planGetSlice, planGetTask)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Printf("id:     %s\n", t.ID)
			fmt.Printf("status: %s\n", t.Status)
			fmt.Println()
			body := t.Task
			if body == "" {
				body = t.Description
			}
			fmt.Println(strings.TrimSpace(body))
		}
		return nil
	},
}
