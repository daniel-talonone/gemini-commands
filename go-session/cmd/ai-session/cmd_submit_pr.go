package main

import (
	"fmt"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/github"
	"github.com/daniel-talonone/gemini-commands/internal/pr"
	"github.com/daniel-talonone/gemini-commands/internal/status"
	"github.com/spf13/cobra"
)

var submitPrCmd = &cobra.Command{
	Use:   "submit-pr [story-id]",
	Short: "Submits a GitHub PR for the given story",
	Long:  `Submits a GitHub pull request. It reads the PR description from pr.md in the feature directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		storyId := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error: could not get current working directory: %v\n", err)
			os.Exit(1)
		}

		remoteURL := git.RemoteURL()

		featureDir, err := feature.ResolveFeatureDir(storyId, cwd, remoteURL)
		if err != nil {
			fmt.Printf("Error: could not resolve feature directory for story '%s': %v\n", storyId, err)
			os.Exit(1)
		}

		s, err := status.LoadStatus(featureDir)
		if err != nil {
			fmt.Printf("Error: could not load status for story '%s': %v\n", storyId, err)
			os.Exit(1)
		}

		prBody, err := pr.Read(featureDir)
		if err != nil {
			fmt.Printf("Error: could not read PR description: %v\n", err)
			os.Exit(1)
		}

		baseBranch := git.DefaultBranch()

		title := "feat: " + s.Branch
		prURL, err := github.CreatePR(s.WorkDir, baseBranch, s.Branch, title, prBody)
		if err != nil {
			fmt.Printf("Error: failed to create PR: %v\n", err)
			os.Exit(1)
		}

		if err := status.WritePRURL(featureDir, prURL); err != nil {
			fmt.Printf("Error: failed to write PR URL to status: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Pull request submitted successfully: %s\n", prURL)
	},
}

func init() {
	rootCmd.AddCommand(submitPrCmd)
}
