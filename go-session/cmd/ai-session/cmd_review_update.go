package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/review"
	"github.com/spf13/cobra"
)

var (
	reviewUpdateFlagJSON    string
	reviewUpdateFlagRegular bool
	reviewUpdateFlagDocs    bool
	reviewUpdateFlagDevOps  bool
)

func init() {
	reviewUpdateCmd.Flags().StringVar(&reviewUpdateFlagJSON, "json", "", "JSON payload of finding updates (required)")
	_ = reviewUpdateCmd.MarkFlagRequired("json")

	reviewUpdateCmd.Flags().BoolVar(&reviewUpdateFlagRegular, "regular", false, "Update regular review findings (default)")
	reviewUpdateCmd.Flags().BoolVar(&reviewUpdateFlagDocs, "docs", false, "Update docs review findings")
	reviewUpdateCmd.Flags().BoolVar(&reviewUpdateFlagDevOps, "devops", false, "Update devops review findings")
	reviewUpdateCmd.MarkFlagsMutuallyExclusive("regular", "docs", "devops")

	rootCmd.AddCommand(reviewUpdateCmd)
}

var reviewUpdateCmd = &cobra.Command{
	Use:   "review-update <story-id>",
	Short: "Atomically update the status and notes of multiple review findings.",
	Long: `Updates the status and notes for multiple review findings in a single atomic operation.
The command reads a JSON payload and applies the updates to the corresponding review file
(review.yml, review-docs.yml, or review-devops.yml).

Arguments:
  <story-id>      The ID of the user story (e.g., "sc-12345") to resolve the feature directory.

Flags:
  --json          Stringified JSON array of update objects.
                  Each object must contain:
                  - "id" (string, required, kebab-case)
                  - "status" (string, required, one of "resolved" or "skipped")
                  - "notes" (string, optional)
  --regular       Update the regular review file (review.yml). This is the default.
  --docs          Update the documentation review file (review-docs.yml).
  --devops        Update the DevOps review file (review-devops.yml).

Example:
  ai-session review-update sc-12345 --json '[{"id":"my-finding","status":"resolved","notes":"Fixed it."}]'
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		storyID := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		remoteURL := git.RemoteURL()

		featureDir, err := feature.ResolveFeatureDir(storyID, cwd, remoteURL)
		if err != nil {
			return fmt.Errorf("failed to resolve feature directory for story %q: %w", storyID, err)
		}

		reviewType := parseReviewUpdateType()

		var updates []review.UpdateRequest
		if err := json.Unmarshal([]byte(reviewUpdateFlagJSON), &updates); err != nil {
			return fmt.Errorf("invalid --json payload: %w. Ensure it is a valid JSON array", err)
		}

		if err := review.UpdateStatuses(featureDir, reviewType, updates); err != nil {
			return err
		}

		fmt.Printf("%d finding(s) updated successfully in %s review file.\n", len(updates), string(reviewType))
		return nil
	},
}

// parseReviewUpdateType determines the review type from the command-line flags.
// It defaults to "regular" if no specific flag is provided.
func parseReviewUpdateType() review.Type {
	if reviewUpdateFlagDocs {
		return review.TypeDocs
	}
	if reviewUpdateFlagDevOps {
		return review.TypeDevOps
	}
	// Default to regular, whether the flag is present or not.
	return review.TypeDefault
}
