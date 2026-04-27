package main

import (
	"fmt"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/spf13/cobra"
)

var (
	getArchitecture bool
	getQuestions    bool
)

var planArtifactsCmd = &cobra.Command{
	Use:   "get <story-id>",
	Short: "Get feature plan artifacts (architecture.md, questions.yml)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		storyID := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current working directory: %w", err)
		}
		remoteURL := git.RemoteURL()
		featureDir, err := feature.ResolveFeatureDir(storyID, cwd, remoteURL)
		if err != nil {
			return fmt.Errorf("failed to resolve feature directory for story ID %s: %w", storyID, err)
		}

		if _, err := os.Stat(featureDir); os.IsNotExist(err) {
			return fmt.Errorf("feature directory does not exist for story ID %s at %s", storyID, featureDir)
		}

		if getArchitecture {
			content, err := plan.LoadArchitecture(featureDir)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), content)
			return err
		}
		if getQuestions {
			content, err := plan.LoadQuestions(featureDir)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), content)
			return err
		}
		return fmt.Errorf("one of --architecture or --questions must be specified")
	},
}

func init() {
	planCmd.AddCommand(planArtifactsCmd)

	planArtifactsCmd.Flags().BoolVar(&getArchitecture, "architecture", false, "Get architecture.md content")
	planArtifactsCmd.Flags().BoolVar(&getQuestions, "questions", false, "Get questions.yml content")
	planArtifactsCmd.MarkFlagsMutuallyExclusive("architecture", "questions")
}
