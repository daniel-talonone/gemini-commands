package main

import (
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "PR operations",
}

func init() {
	rootCmd.AddCommand(prCmd)
}
