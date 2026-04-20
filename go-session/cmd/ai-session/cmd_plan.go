package main

import "github.com/spf13/cobra"

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Operations for planning a feature.",
	Long:  `A group of commands for creating, updating, and managing the implementation plan for a feature.`,
}

func init() {
	rootCmd.AddCommand(planCmd)
}
