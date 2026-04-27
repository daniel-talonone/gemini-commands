package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/daniel-talonone/gemini-commands/internal/status"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type loadContentFunc func(featureDir string) (string, error)

// setupTestCommand creates a test Cobra command structure for the `ai-session plan get` command.
// It uses mocked file loading functions to isolate the command's flag parsing and output logic
// from actual file system interactions. This means it doesn't provide full end-to-end verification
// of `feature.ResolveFeatureDir` or the complete command execution flow. The feature directory
// path is constructed directly for testing purposes.
func setupTestCommand(aisessionHome string,
	loadArchitecture loadContentFunc, loadQuestions loadContentFunc) (*cobra.Command, *bytes.Buffer) {
	rootCmd := &cobra.Command{Use: "ai-session"}
	planCmd := &cobra.Command{Use: "plan"}
	rootCmd.AddCommand(planCmd)

	// Local flags to be used by the RunE closure
	var getArchitecture bool
	var getQuestions bool

	planArtifactsCmd := &cobra.Command{
		Use:   "get <story-id>",
		Short: "Get feature plan artifacts (architecture.md, questions.yml)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			storyID := args[0]

			// Construct the feature directory path directly for testing
			featureDir := filepath.Join(aisessionHome, ".features", "test", "repo", storyID)

			// Temporarily set AI_SESSION_HOME for the duration of this RunE execution
			// This is important because feature.ResolveFeatureDir relies on it when
			// it is called in the real command. For the test, we bypass it.
			originalAisessionHome := os.Getenv("AI_SESSION_HOME")
			_ = os.Setenv("AI_SESSION_HOME", aisessionHome)
			defer func() { _ = os.Setenv("AI_SESSION_HOME", originalAisessionHome) }()

			if _, err := os.Stat(featureDir); os.IsNotExist(err) {
				return fmt.Errorf("feature directory does not exist for story ID %s at %s", storyID, featureDir)
			}

			if getArchitecture {
				content, err := loadArchitecture(featureDir)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), content)
				return err
			}
			if getQuestions {
				content, err := loadQuestions(featureDir)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), content)
				return err
			}
			return fmt.Errorf("one of --architecture or --questions must be specified")
		},
	}
	planArtifactsCmd.Flags().BoolVar(&getArchitecture, "architecture", false, "Get architecture.md content")
	planArtifactsCmd.Flags().BoolVar(&getQuestions, "questions", false, "Get questions.yml content")
	planArtifactsCmd.MarkFlagsMutuallyExclusive("architecture", "questions")

	planCmd.AddCommand(planArtifactsCmd)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	return rootCmd, buf
}

func TestPlanGetQuestions(t *testing.T) {
	t.Run("returns content when questions.yml exists", func(t *testing.T) {
		tempAisessionHome := t.TempDir()
		rootCmd, buf := setupTestCommand(tempAisessionHome, plan.LoadArchitecture, plan.LoadQuestions)

		storyID := "test-story-questions-exist"
		dummyFeatureDir := filepath.Join(tempAisessionHome, ".features", "test", "repo", storyID)
		require.NoError(t, os.MkdirAll(dummyFeatureDir, 0755))
		require.NoError(t, status.Create(dummyFeatureDir, "test/repo", "test-branch", "/work", "", ""))

		expectedContent := `questions:
  - id: q1
    question: "What is the meaning of life?"
    status: open
`
		require.NoError(t, os.WriteFile(filepath.Join(dummyFeatureDir, "questions.yml"), []byte(expectedContent), 0644))

		rootCmd.SetArgs([]string{"plan", "get", "--questions", storyID})
		err := rootCmd.Execute()
		require.NoError(t, err)
		assert.Equal(t, expectedContent, buf.String())
	})

	t.Run("returns empty output without error when questions.yml does not exist", func(t *testing.T) {
		tempAisessionHome := t.TempDir()
		rootCmd, buf := setupTestCommand(tempAisessionHome, plan.LoadArchitecture, plan.LoadQuestions)

		storyID := "test-story-questions-not-exist"
		dummyFeatureDir := filepath.Join(tempAisessionHome, ".features", "test", "repo", storyID)
		require.NoError(t, os.MkdirAll(dummyFeatureDir, 0755))
		require.NoError(t, status.Create(dummyFeatureDir, "test/repo", "test-branch", "/work", "", ""))

		rootCmd.SetArgs([]string{"plan", "get", "--questions", storyID})
		err := rootCmd.Execute()
		require.NoError(t, err)
		assert.Empty(t, buf.String())
	})

	t.Run("returns error when feature directory does not exist", func(t *testing.T) {
		tempAisessionHome := t.TempDir()
		rootCmd, buf := setupTestCommand(tempAisessionHome, plan.LoadArchitecture, plan.LoadQuestions)

		storyID := "non-existent-feature-dir"
		// Do not create the dummyFeatureDir

		rootCmd.SetArgs([]string{"plan", "get", "--questions", storyID})
		err := rootCmd.Execute()
		require.Error(t, err)
		assert.Contains(t, buf.String(), fmt.Sprintf("feature directory does not exist for story ID %s", storyID))
	})

	t.Run("flags --architecture and --questions are mutually exclusive", func(t *testing.T) {
		tempAisessionHome := t.TempDir()
		rootCmd, buf := setupTestCommand(tempAisessionHome, plan.LoadArchitecture, plan.LoadQuestions)

		storyID := "any-story-id"
		rootCmd.SetArgs([]string{"plan", "get", "--architecture", "--questions", storyID})
		err := rootCmd.Execute()
		require.Error(t, err)
		assert.Contains(t, buf.String(), "if any flags in the group [architecture questions] are set none of the others can be")
	})
}
