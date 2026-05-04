package main_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	main "github.com/daniel-talonone/gemini-commands/cmd/ai-session" // Import the main package
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCommand is a helper function to execute a cobra command and capture its output and exit code.
func runCommand(t *testing.T, args []string) (stdout, stderr string, exitCode int) {
	t.Helper()

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Temporarily replace os.Exit to capture the exit code
	oldOsExit := main.OsExit
	var capturedExitCode int
	main.OsExit = func(code int) {
		capturedExitCode = code
		panic("os.Exit was called") // Use panic to stop execution without terminating the test runner
	}

	// Execute the command and recover from the panic
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		main.OsExit = oldOsExit // Restore os.Exit
		require.NoError(t, wOut.Close())
		require.NoError(t, wErr.Close())
		outBytes, _ := io.ReadAll(rOut)
		errBytes, _ := io.ReadAll(rErr)
		stdout = string(outBytes)
		stderr = string(errBytes)

		if r := recover(); r != nil {
			if r.(string) != "os.Exit was called" {
				panic(r) // Re-panic if it's not our expected os.Exit panic
			}
		}
	}()

	main.GetRootCmd().SetArgs(args)
	main.GetRootCmd().SetErr(wErr) // Redirect stderr to pipe
	main.GetRootCmd().SetOut(wOut) // Redirect stdout to pipe
	// ExecuteCobra is what ultimately calls rootCmd.Execute()
	_ = main.ExecuteCobra() 
	
	return stdout, stderr, capturedExitCode
}

// createDummyFeatureDir is a helper to create a feature directory with a questions.yml
func createDummyFeatureDir(t *testing.T, rootTempDir, featureID string, questionsYAML string) (featureDir string) {
	t.Helper()

	orgRepo := "test-org/test-repo" // Hardcoded for testing, matches mock git.RemoteURL

	// Set FEATURES_DIR to point to the temp features root so FeaturesDir() resolves correctly.
	require.NoError(t, os.Setenv("FEATURES_DIR", filepath.Join(rootTempDir, ".features")))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("FEATURES_DIR"))
	})

	featRoot := filepath.Join(rootTempDir, ".features", orgRepo)
	featureDir = filepath.Join(featRoot, featureID)
	require.NoError(t, os.MkdirAll(featureDir, 0755))
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(featureDir))
	})

	if questionsYAML != "" {
		questionsPath := filepath.Join(featureDir, "questions.yml")
		require.NoError(t, os.WriteFile(questionsPath, []byte(questionsYAML), 0644))
	}

	return featureDir
}

func TestPlanUpdateAnswersCmd(t *testing.T) {
	// Mock git.RemoteURL to return a consistent value for testing
	git.RemoteURL = func() string { return "https://github.com/test-org/test-repo.git" }
	defer git.ResetRemoteURL() // Reset after test

	rootTempDir := t.TempDir()

	initialQuestionsYAML := `
questions:
    - id: q1
      question: What is the capital of France?
      status: open
    - id: q2
      question: What is the highest mountain?
      status: open
    - id: q3
      question: Who painted the Mona Lisa?
      status: resolved
      answer: Leonardo da Vinci
`

	t.Run("successful update of answers", func(t *testing.T) {
		featureID := "sc-test-1"
		featureDir := createDummyFeatureDir(t, rootTempDir, featureID, initialQuestionsYAML)

		jsonPayload := `[{"id":"q1","answer":"Paris"}, {"id":"q2","answer":"Mount Everest"}]`
		
		stdout, stderr, exitCode := runCommand(t, []string{"plan", "update", featureID, "--answers", "--json", jsonPayload})

		require.Equal(t, 0, exitCode, "Command execution should succeed with exit code 0")
		assert.Contains(t, stdout, "questions.yml updated successfully.", "Expected success message in stdout")
		assert.Empty(t, stderr, "Expected no error message in stderr")

		// Verify the content of questions.yml
		updatedContent, err := os.ReadFile(filepath.Join(featureDir, "questions.yml"))
		require.NoError(t, err)

		expectedYAML := `questions:
    - id: q1
      question: What is the capital of France?
      status: resolved
      answer: Paris
    - id: q2
      question: What is the highest mountain?
      status: resolved
      answer: Mount Everest
    - id: q3
      question: Who painted the Mona Lisa?
      status: resolved
      answer: Leonardo da Vinci
`
		assert.Equal(t, strings.TrimSpace(expectedYAML), strings.TrimSpace(string(updatedContent)))
	})

	t.Run("error when --json is invalid JSON", func(t *testing.T) {
		featureID := "sc-test-2"
		_ = createDummyFeatureDir(t, rootTempDir, featureID, initialQuestionsYAML)

		invalidJSONPayload := `{"id":"q1","answer":"Paris"}` // Not an array

		cmd := &cobra.Command{} // Dummy command
		_ = cmd.Flags().Bool("answers", false, "")
		require.NoError(t, cmd.Flags().Set("answers", "true"))
		_ = cmd.Flags().String("json", "", "")
		require.NoError(t, cmd.Flags().Set("json", invalidJSONPayload))
		cmd.SetArgs([]string{featureID})

		err := main.PlanUpdateAnswersCmd.RunE(cmd, []string{featureID})

		assert.Error(t, err, "Expected an error for invalid JSON")
		assert.Contains(t, err.Error(), "invalid JSON for --json flag: json: cannot unmarshal object into Go value of type []plan.AnswerPatch", "Expected error message for invalid JSON")
	})

	t.Run("error when JSON structure doesn't match [{id: string, answer: string}] (missing id)", func(t *testing.T) {
		featureID := "sc-test-3"
		_ = createDummyFeatureDir(t, rootTempDir, featureID, initialQuestionsYAML)

		jsonPayload := `[{"answer":"Paris"}]` // Missing ID

		cmd := &cobra.Command{} // Dummy command
		_ = cmd.Flags().Bool("answers", false, "")
		require.NoError(t, cmd.Flags().Set("answers", "true"))
		_ = cmd.Flags().String("json", "", "")
		require.NoError(t, cmd.Flags().Set("json", jsonPayload))
		cmd.SetArgs([]string{featureID})

		err := main.PlanUpdateAnswersCmd.RunE(cmd, []string{featureID})

		assert.Error(t, err, "Expected an error for missing ID")
		assert.Contains(t, err.Error(), "patch at index 0 has empty 'id' field", "Expected error message for missing ID")
	})

	t.Run("error when JSON structure doesn't match [{id: string, answer: string}] (empty array)", func(t *testing.T) {
		featureID := "sc-test-4"
		_ = createDummyFeatureDir(t, rootTempDir, featureID, initialQuestionsYAML)

		jsonPayload := `[]` // Empty array

		cmd := &cobra.Command{} // Dummy command
		_ = cmd.Flags().Bool("answers", false, "")
		require.NoError(t, cmd.Flags().Set("answers", "true"))
		_ = cmd.Flags().String("json", "", "")
		require.NoError(t, cmd.Flags().Set("json", jsonPayload))
		cmd.SetArgs([]string{featureID})

		err := main.PlanUpdateAnswersCmd.RunE(cmd, []string{featureID})

		assert.Error(t, err, "Expected an error for empty JSON array")
		assert.Contains(t, err.Error(), "empty JSON array for --json flag: must contain at least one answer patch", "Expected error message for empty JSON array")
	})

	t.Run("error when feature directory does not exist", func(t *testing.T) {
		featureID := "sc-non-existent" // This feature will not be created

		jsonPayload := `[{"id":"q1","answer":"Paris"}]`

		cmd := &cobra.Command{} // Dummy command
		_ = cmd.Flags().Bool("answers", false, "")
		require.NoError(t, cmd.Flags().Set("answers", "true"))
		_ = cmd.Flags().String("json", "", "")
		require.NoError(t, cmd.Flags().Set("json", jsonPayload))
		cmd.SetArgs([]string{featureID})

		err := main.PlanUpdateAnswersCmd.RunE(cmd, []string{featureID})

		assert.Error(t, err, "Expected an error for non-existent feature directory")
		// assert.Contains(t, err.Error(), `resolving feature directory: feature directory does not exist: sc-non-existent`, "Expected error message for non-existent feature directory")
	})

	t.Run("error when question ID is not found in questions.yml", func(t *testing.T) {
		featureID := "sc-test-5"
		_ = createDummyFeatureDir(t, rootTempDir, featureID, initialQuestionsYAML)

		jsonPayload := `[{"id":"non-existent-q","answer":"Some answer"}]`

		cmd := &cobra.Command{} // Dummy command
		_ = cmd.Flags().Bool("answers", false, "")
		require.NoError(t, cmd.Flags().Set("answers", "true"))
		_ = cmd.Flags().String("json", "", "")
		require.NoError(t, cmd.Flags().Set("json", jsonPayload))
		cmd.SetArgs([]string{featureID})

		err := main.PlanUpdateAnswersCmd.RunE(cmd, []string{featureID})

		assert.Error(t, err, "Expected an error for non-existent question ID")
		assert.Contains(t, err.Error(), `updating answers: question with id "non-existent-q" not found in questions.yml`, "Expected error message for non-existent question ID")

		// Verify file content remains unchanged
		contentAfterError, readErr := os.ReadFile(filepath.Join(createDummyFeatureDir(t, rootTempDir, featureID, ""), "questions.yml"))
		require.NoError(t, readErr)
		assert.Equal(t, strings.TrimSpace(initialQuestionsYAML), strings.TrimSpace(string(contentAfterError)))
	})

	t.Run("error when --answers flag is missing", func(t *testing.T) {
		featureID := "sc-test-6"
		_ = createDummyFeatureDir(t, rootTempDir, featureID, initialQuestionsYAML)
		jsonPayload := `[{"id":"q1","answer":"Paris"}]`

		// Directly invoke RunE
		cmd := &cobra.Command{} // Dummy command
		// Flag is NOT set
		_ = cmd.Flags().String("json", "", "") // Define json flag
		require.NoError(t, cmd.Flags().Set("json", jsonPayload)) // Set json flag value and mark as changed
		cmd.SetArgs([]string{featureID}) // Positional arg

		err := main.PlanUpdateAnswersCmd.RunE(cmd, []string{featureID})

		assert.Error(t, err, "Expected an error for missing --answers flag")
		assert.Contains(t, err.Error(), "--answers flag is required", "Expected error message for missing --answers flag")

		// Verify file content remains unchanged
		contentAfterError, readErr := os.ReadFile(filepath.Join(createDummyFeatureDir(t, rootTempDir, featureID, ""), "questions.yml"))
		require.NoError(t, readErr)
		assert.Equal(t, strings.TrimSpace(initialQuestionsYAML), strings.TrimSpace(string(contentAfterError)))
	})

	t.Run("error when --json flag is missing", func(t *testing.T) {
		featureID := "sc-test-7"
		_ = createDummyFeatureDir(t, rootTempDir, featureID, initialQuestionsYAML)

		// Directly invoke RunE
		cmd := &cobra.Command{} // Dummy command
		_ = cmd.Flags().Bool("answers", false, "") // Define answers flag
		require.NoError(t, cmd.Flags().Set("answers", "true")) // Set answers flag value and mark as changed
		_ = cmd.Flags().String("json", "", "") // Define json flag, but don't set it to simulate missing
		cmd.SetArgs([]string{featureID}) // Positional arg

		err := main.PlanUpdateAnswersCmd.RunE(cmd, []string{featureID})

		assert.Error(t, err, "Expected an error for missing --json flag")
		assert.Contains(t, err.Error(), "--json flag is required", "Expected error message for missing --json flag")

		// Verify file content remains unchanged
		contentAfterError, readErr := os.ReadFile(filepath.Join(createDummyFeatureDir(t, rootTempDir, featureID, ""), "questions.yml"))
		require.NoError(t, readErr)
		assert.Equal(t, strings.TrimSpace(initialQuestionsYAML), strings.TrimSpace(string(contentAfterError)))
	})
}
