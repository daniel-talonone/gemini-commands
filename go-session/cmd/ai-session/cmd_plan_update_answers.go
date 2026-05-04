package main

import (
	"encoding/json"
	"fmt"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/spf13/cobra"
)

var PlanUpdateAnswersCmd = &cobra.Command{
	Use:   "update <story-id> --answers --json <json>",
	Short: "Update answers in questions.yml",
	Long: `Update answers in questions.yml for a given story.

This command patches matching questions: sets 'answer' to the provided value and 'status' to "resolved".
If any 'id' in the payload is not found in questions.yml, the command fails.
The '--answers' and '--json' flags are both required.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("answers") {
			return fmt.Errorf("--answers flag is required")
		}

		jsonPayload, err := cmd.Flags().GetString("json")
		if err != nil {
			return fmt.Errorf("getting --json flag: %w", err)
		}
		if !cmd.Flags().Changed("json") { // Explicitly check if the flag was provided
			return fmt.Errorf("--json flag is required")
		}

		storyID := args[0]

		var patches []plan.AnswerPatch
		if err := json.Unmarshal([]byte(jsonPayload), &patches); err != nil {
			return fmt.Errorf("invalid JSON for --json flag: %w", err)
		}

		if len(patches) == 0 {
			return fmt.Errorf("empty JSON array for --json flag: must contain at least one answer patch")
		}

		for i, patch := range patches {
			if patch.ID == "" {
				return fmt.Errorf("patch at index %d has empty 'id' field", i)
			}
			if patch.Answer == "" {
				return fmt.Errorf("patch at index %d with id '%s' has empty 'answer' field", i, patch.ID)
			}
		}

		featureDir, err := feature.ResolveFeatureDir(storyID, ".", git.RemoteURL())
		if err != nil {
			return fmt.Errorf("resolving feature directory: %w", err)
		}

		if err := plan.UpdateAnswers(featureDir, patches); err != nil {
			return fmt.Errorf("updating answers: %w", err)
		}

		fmt.Println("questions.yml updated successfully.")
		return nil
	},
}

func init() {
	planCmd.AddCommand(PlanUpdateAnswersCmd)

	PlanUpdateAnswersCmd.Flags().Bool("answers", false, "Required flag to indicate answers are being updated.")
	PlanUpdateAnswersCmd.Flags().String("json", "", "Required flag. A JSON string representing an array of answer patches.")


}
