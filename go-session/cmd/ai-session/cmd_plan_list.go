package main

import (
	"fmt"
	"os"
	"strings"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/spf13/cobra"
)

var planListSlice string

func init() {
	planListCmd.Flags().StringVar(&planListSlice, "slice", "", "List tasks within this slice ID instead of listing slices")
	rootCmd.AddCommand(planListCmd)
}

var planListCmd = &cobra.Command{
	Use:   "plan-list <feature-dir>",
	Short: "List slices (or tasks within a slice) from plan.yml",
	Long: `Lists slices or tasks from plan.yml without loading the full file content.

Without --slice: prints every slice ID and its status, one per line.
With --slice <id>: prints every task ID and its status within that slice.

Arguments:
  <feature-dir>  Path to the feature directory containing plan.yml

Flags:
  --slice  Slice ID to list tasks for (optional)

Output format (one entry per line):
  <id>  [<status>]  <description-or-task-preview>

Errors:
  - plan.yml not found in <feature-dir>
  - Slice not found (when --slice is given)
  - Exactly 1 positional argument required`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 argument required: <feature-dir>, got %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		featureDir := args[0]

		if planListSlice == "" {
			slices, err := commands.ListSlices(featureDir)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			for _, s := range slices {
				deps := ""
				if len(s.DependsOn) > 0 {
					deps = "  depends_on: " + strings.Join(s.DependsOn, ", ")
				}
				fmt.Printf("%-40s [%s]%s\n", s.ID, s.Status, deps)
				if s.Description != "" {
					fmt.Printf("  %s\n", s.Description)
				}
			}
		} else {
			tasks, err := commands.ListTasks(featureDir, planListSlice)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			for _, t := range tasks {
				fmt.Printf("%-40s [%s]\n", t.ID, t.Status)
			}
		}
		return nil
	},
}
