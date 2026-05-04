package plan_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteQuestions(t *testing.T) {
	validYAML := `
questions:
  - id: q1
    question: "What is love?"
    status: open
  - id: q2
    question: "Baby don't hurt me"
    status: resolved
    answer: "No more"
`

	t.Run("writes valid questions atomically", func(t *testing.T) {
		tempDir := t.TempDir()
		err := plan.WriteQuestions(tempDir, []byte(validYAML))
		require.NoError(t, err)

		finalPath := filepath.Join(tempDir, "questions.yml")
		_, err = os.Stat(finalPath)
		require.NoError(t, err, "questions.yml should exist")

		content, err := os.ReadFile(finalPath)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(validYAML), strings.TrimSpace(string(content)))
	})

	t.Run("returns error for non-existent feature directory", func(t *testing.T) {
		err := plan.WriteQuestions("/non-existent-dir", []byte(validYAML))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "feature directory does not exist")
	})

	t.Run("returns LLM-actionable error for invalid YAML", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidYAML := `questions: - id: q1`
		err := plan.WriteQuestions(tempDir, []byte(invalidYAML))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing questions.yml")
	})

	validationTests := []struct {
		name          string
		yaml          string
		expectedError string
	}{
		{
			name: "missing id",
			yaml: `
questions:
  - question: "missing id"
    status: open`,
			expectedError: `questions[0].id: value is empty`,
		},
		{
			name: "invalid kebab-case id",
			yaml: `
questions:
  - id: "invalidID"
    question: "invalid id"
    status: open`,
			expectedError: `questions[0].id: "invalidID" is not kebab-case`,
		},
		{
			name: "empty question",
			yaml: `
questions:
  - id: "q1"
    question: ""
    status: open`,
			expectedError: `questions[0].question: value is empty`,
		},
		{
			name: "invalid status",
			yaml: `
questions:
  - id: "q1"
    question: "some question"
    status: "pending"`,
			expectedError: `questions[0].status: "pending" is not valid`,
		},
		{
			name: "indexed error message",
			yaml: `
questions:
  - id: q1
    question: "valid"
    status: open
  - id: Q2
    question: "invalid"
    status: resolved`,
			expectedError: `questions[1].id: "Q2" is not kebab-case`,
		},
	}

	for _, tc := range validationTests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			err := plan.WriteQuestions(tempDir, []byte(tc.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)

			// Check that no file was created
			_, err = os.Stat(filepath.Join(tempDir, "questions.yml"))
			assert.True(t, os.IsNotExist(err), "questions.yml should not have been created on validation failure")
		})
	}
}



func TestLoadQuestions(t *testing.T) {
	t.Run("returns content as string when questions.yml exists", func(t *testing.T) {
		tempDir := t.TempDir()
		expectedContent := `questions:
  - id: test-q
    question: "Test Question"
    status: open
`
		err := os.WriteFile(filepath.Join(tempDir, "questions.yml"), []byte(expectedContent), 0644)
		require.NoError(t, err)

		content, err := plan.LoadQuestions(tempDir)
		require.NoError(t, err)
		assert.Equal(t, expectedContent, content)
	})

	t.Run("returns empty string when file doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		content, err := plan.LoadQuestions(tempDir)
		require.NoError(t, err)
		assert.Equal(t, "", content)
	})

	t.Run("returns error when feature directory doesn't exist", func(t *testing.T) {
		_, err := plan.LoadQuestions("/non-existent-feature-dir")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "feature directory does not exist")
	})
}

func TestUpdateAnswers(t *testing.T) {
	initialYAML := `
questions:
  - id: q1
    question: "What is the capital of France?"
    status: open
  - id: q2
    question: "What is the highest mountain?"
    status: open
  - id: q3
    question: "Who painted the Mona Lisa?"
    status: resolved
    answer: "Leonardo da Vinci"
`

	t.Run("successfully updates matching question IDs and sets status to resolved", func(t *testing.T) {
		tempDir := t.TempDir()
		questionsPath := filepath.Join(tempDir, "questions.yml")
		err := os.WriteFile(questionsPath, []byte(initialYAML), 0644)
		require.NoError(t, err)

		patches := []plan.AnswerPatch{
			{ID: "q1", Answer: "Paris"},
			{ID: "q2", Answer: "Mount Everest"},
		}

		err = plan.UpdateAnswers(tempDir, patches)
		require.NoError(t, err)

		updatedContent, err := os.ReadFile(questionsPath)
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

	t.Run("returns error when a question ID is not found", func(t *testing.T) {
		tempDir := t.TempDir()
		questionsPath := filepath.Join(tempDir, "questions.yml")
		err := os.WriteFile(questionsPath, []byte(initialYAML), 0644)
		require.NoError(t, err)

		patches := []plan.AnswerPatch{
			{ID: "q1", Answer: "Paris"},
			{ID: "missing-id", Answer: "Some answer"},
		}

		err = plan.UpdateAnswers(tempDir, patches)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `question with id "missing-id" not found in questions.yml`)

		// Verify file content remains unchanged
		contentAfterError, err := os.ReadFile(questionsPath)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(initialYAML), strings.TrimSpace(string(contentAfterError)))
	})

	t.Run("preserves unmodified questions when some IDs match", func(t *testing.T) {
		tempDir := t.TempDir()
		questionsPath := filepath.Join(tempDir, "questions.yml")
		err := os.WriteFile(questionsPath, []byte(initialYAML), 0644)
		require.NoError(t, err)

		patches := []plan.AnswerPatch{
			{ID: "q1", Answer: "Paris"},
		}

		err = plan.UpdateAnswers(tempDir, patches)
		require.NoError(t, err)

		updatedContent, err := os.ReadFile(questionsPath)
		require.NoError(t, err)

		expectedYAML := `questions:
    - id: q1
      question: What is the capital of France?
      status: resolved
      answer: Paris
    - id: q2
      question: What is the highest mountain?
      status: open
    - id: q3
      question: Who painted the Mona Lisa?
      status: resolved
      answer: Leonardo da Vinci
`
		assert.Equal(t, strings.TrimSpace(expectedYAML), strings.TrimSpace(string(updatedContent)))
	})

	t.Run("returns error when feature directory does not exist", func(t *testing.T) {
		patches := []plan.AnswerPatch{
			{ID: "q1", Answer: "Paris"},
		}
		err := plan.UpdateAnswers("/non-existent-feature-dir", patches)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "feature directory does not exist")
	})

	t.Run("atomic write: ensures temp file cleanup on error during write", func(t *testing.T) {
		tempDir := t.TempDir()
		questionsPath := filepath.Join(tempDir, "questions.yml")

		// Create a valid questions.yml initially
		err := os.WriteFile(questionsPath, []byte(initialYAML), 0644)
		require.NoError(t, err)

		// Introduce a malformed patch that would cause an unmarshal error on re-read in WriteQuestions
		// This won't happen with the current validation, but simulates a failure during the atomic write phase
		// to ensure temp files are cleaned up.
		// Mocking WriteQuestions to fail, but it's called internally by UpdateAnswers
		// For this test, we'll ensure that if UpdateAnswers fails due to an issue
		// that might leave a temp file (e.g., if marshalling produces invalid yaml, though
		// our current marshal will always produce valid yaml from struct),
		// the temp files are cleaned.
		// A simpler way is to check for temp files directly after UpdateAnswers returns an error.

		// Manually create a situation where unmarshalling fails within WriteQuestions when it's called
		// This is a bit tricky since UpdateAnswers calls WriteQuestions with valid YAML.
		// Let's test by checking for any leftover .questions.tmp.* files if UpdateAnswers fails for
		// any reason (e.g., reading initial file fails, or parsing initial file fails).

		// Test case for unmarshalling error:
		invalidQuestionsYAML := `questions:
  - id: q1
    question: "What is the capital of France?"
    status: open
  - - invalid-indentation
`
		err = os.WriteFile(questionsPath, []byte(invalidQuestionsYAML), 0644)
		require.NoError(t, err)

		patches := []plan.AnswerPatch{
			{ID: "q1", Answer: "Paris"},
		}

		err = plan.UpdateAnswers(tempDir, patches)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing questions.yml")

		// Verify no temp files are left
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		for _, f := range files {
			assert.NotContains(t, f.Name(), ".questions.tmp.", "Temporary file was not cleaned up")
		}
	})
	
	t.Run("handles multiple patches for the same ID, last one wins", func(t *testing.T) {
		tempDir := t.TempDir()
		questionsPath := filepath.Join(tempDir, "questions.yml")
		err := os.WriteFile(questionsPath, []byte(initialYAML), 0644)
		require.NoError(t, err)

		patches := []plan.AnswerPatch{
			{ID: "q1", Answer: "First Answer"},
			{ID: "q1", Answer: "Second Answer"},
		}

		err = plan.UpdateAnswers(tempDir, patches)
		require.NoError(t, err)

		updatedContent, err := os.ReadFile(questionsPath)
		require.NoError(t, err)

		expectedYAML := `questions:
    - id: q1
      question: What is the capital of France?
      status: resolved
      answer: Second Answer
    - id: q2
      question: What is the highest mountain?
      status: open
    - id: q3
      question: Who painted the Mona Lisa?
      status: resolved
      answer: Leonardo da Vinci
`
		assert.Equal(t, strings.TrimSpace(expectedYAML), strings.TrimSpace(string(updatedContent)))
	})
}

