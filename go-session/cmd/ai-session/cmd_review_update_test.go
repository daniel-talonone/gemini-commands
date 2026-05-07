package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/review"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReviewUpdateCmd_HappyPath tests the successful update of a review finding.
func TestReviewUpdateCmd_HappyPath(t *testing.T) {
	// Setup: Create a temporary feature directory with a review file.
	dir := t.TempDir()
	featureDir := filepath.Join(dir, "sc-123")
	require.NoError(t, os.MkdirAll(featureDir, 0755))
	require.NoError(t, review.Write(featureDir, review.TypeDefault, []review.Finding{
		{ID: "test-finding", Status: "open", Feedback: "initial feedback"},
	}))

	// Mock the feature directory resolution.
	originalResolveFeatureDir := feature.ResolveFeatureDirImpl
	feature.ResolveFeatureDirImpl = func(storyID, cwd, remoteURL string) (string, error) {
		return featureDir, nil
	}
	defer func() { feature.ResolveFeatureDirImpl = originalResolveFeatureDir }()

	// Capture stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command.
	cmd := rootCmd
	cmd.SetArgs([]string{"review-update", "sc-123", "--json", `[{"id":"test-finding","status":"resolved","notes":"fixed"}]`})
	err := cmd.Execute()

	// Restore stdout and read the output.
	require.NoError(t, w.Close())
	os.Stdout = oldStdout
	var out bytes.Buffer
	_, _ = out.ReadFrom(r)

	// Assertions.
	require.NoError(t, err, "Command should execute without error")
	assert.Contains(t, out.String(), "1 finding(s) updated successfully", "Output should confirm update")

	// Verify the file content.
	findings, err := review.Load(featureDir, review.TypeDefault)
	require.NoError(t, err, "Should be able to load findings after update")
	require.Len(t, findings, 1, "There should be one finding")
	assert.Equal(t, "resolved", findings[0].Status, "Status should be updated to resolved")
	assert.Equal(t, "fixed", findings[0].Notes, "Notes should be updated")
}

// TestReviewUpdateCmd_Error_InvalidJSON tests the command's failure on malformed JSON.
func TestReviewUpdateCmd_Error_InvalidJSON(t *testing.T) {
	cmd := rootCmd
	cmd.SetArgs([]string{"review-update", "sc-123", "--json", "not-json"})
	err := cmd.Execute()

	require.Error(t, err, "Command should fail with invalid JSON")
	assert.Contains(t, err.Error(), "invalid --json payload", "Error message should indicate a JSON issue")
}

// TestReviewUpdateCmd_Error_FindingNotFound tests failure when a finding ID is not found.
func TestReviewUpdateCmd_Error_FindingNotFound(t *testing.T) {
	dir := t.TempDir()
	featureDir := filepath.Join(dir, "sc-123")
	require.NoError(t, os.MkdirAll(featureDir, 0755))
	require.NoError(t, review.Write(featureDir, review.TypeDefault, []review.Finding{}))

	originalResolveFeatureDir := feature.ResolveFeatureDirImpl
	feature.ResolveFeatureDirImpl = func(storyID, cwd, remoteURL string) (string, error) {
		return featureDir, nil
	}
	defer func() { feature.ResolveFeatureDirImpl = originalResolveFeatureDir }()

	cmd := rootCmd
	cmd.SetArgs([]string{"review-update", "sc-123", "--json", `[{"id":"non-existent","status":"resolved"}]`})
	err := cmd.Execute()

	require.Error(t, err, "Command should fail if finding ID is not found")
	assert.Contains(t, err.Error(), `finding ID "non-existent" not found`, "Error message should specify the missing ID")
}
