package review

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func validFinding(id string) Finding {
	return Finding{ID: id, File: "main.go", Line: 1, Feedback: "some feedback", Status: "open"}
}

// --- Create ---

func TestCreate_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, TypeDefault))
	_, err := os.Stat(filepath.Join(dir, "review.yml"))
	require.NoError(t, err)
}

func TestCreate_Idempotent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, TypeDefault))
	// Write something so we can detect an overwrite
	path := filepath.Join(dir, "review.yml")
	require.NoError(t, os.WriteFile(path, []byte("# keep me"), 0644))
	require.NoError(t, Create(dir, TypeDefault))
	data, _ := os.ReadFile(path)
	assert.Equal(t, "# keep me", string(data), "second Create must not overwrite existing file")
}

func TestCreate_NoTmpLeftover(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, TypeDefault))
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".tmp", "no temp files should remain after Create")
	}
}

func TestCreate_UnknownType(t *testing.T) {
	dir := t.TempDir()
	err := Create(dir, Type("unknown"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown review type")
}

func TestCreate_DocsType(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, TypeDocs))
	_, err := os.Stat(filepath.Join(dir, "review-docs.yml"))
	require.NoError(t, err)
}

func TestCreate_DevOpsType(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, TypeDevOps))
	_, err := os.Stat(filepath.Join(dir, "review-devops.yml"))
	require.NoError(t, err)
}

// --- Load ---

func TestLoad_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	findings, err := Load(dir, TypeDefault)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestLoad_HappyPath(t *testing.T) {
	dir := t.TempDir()
	f := validFinding("find-1")
	require.NoError(t, Append(dir, TypeDefault, f))
	findings, err := Load(dir, TypeDefault)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "find-1", findings[0].ID)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review.yml")
	require.NoError(t, os.WriteFile(path, []byte(":\tinvalid: yaml: :"), 0644))
	_, err := Load(dir, TypeDefault)
	require.Error(t, err)
}

func TestLoad_InvalidFinding_BadStatus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review.yml")
	require.NoError(t, os.WriteFile(path, []byte("- id: find-1\n  feedback: oops\n  status: invalid\n"), 0644))
	_, err := Load(dir, TypeDefault)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid finding")
}

func TestLoad_InvalidFinding_BadID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review.yml")
	require.NoError(t, os.WriteFile(path, []byte("- id: Not_Valid\n  feedback: oops\n  status: open\n"), 0644))
	_, err := Load(dir, TypeDefault)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid finding")
}

func TestLoad_InvalidFinding_EmptyFeedback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review.yml")
	require.NoError(t, os.WriteFile(path, []byte("- id: find-1\n  feedback: \"\"\n  status: open\n"), 0644))
	_, err := Load(dir, TypeDefault)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid finding")
}

func TestLoad_UnknownType(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir, Type("unknown"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown review type")
}

// --- Append ---

func TestAppend_CreatesFileIfAbsent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Append(dir, TypeDefault, validFinding("find-1")))
	findings, err := Load(dir, TypeDefault)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "find-1", findings[0].ID)
}

func TestAppend_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Append(dir, TypeDefault, validFinding("find-1")))
	require.NoError(t, Append(dir, TypeDefault, validFinding("find-2")))
	findings, err := Load(dir, TypeDefault)
	require.NoError(t, err)
	assert.Len(t, findings, 2)
}

func TestAppend_RejectsEmptyID(t *testing.T) {
	dir := t.TempDir()
	f := validFinding("")
	err := Append(dir, TypeDefault, f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID must not be empty")
	assertNoDefaultReviewFile(t, dir)
}

func TestAppend_RejectsNonKebabID(t *testing.T) {
	dir := t.TempDir()
	f := validFinding("Not_Kebab")
	err := Append(dir, TypeDefault, f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kebab-case")
	assertNoDefaultReviewFile(t, dir)
}

func TestAppend_RejectsEmptyFeedback(t *testing.T) {
	dir := t.TempDir()
	f := Finding{ID: "find-1", Status: "open"}
	err := Append(dir, TypeDefault, f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Feedback must not be empty")
	assertNoDefaultReviewFile(t, dir)
}

func TestAppend_RejectsInvalidStatus(t *testing.T) {
	dir := t.TempDir()
	f := Finding{ID: "find-1", Feedback: "feedback", Status: "todo"}
	err := Append(dir, TypeDefault, f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Status")
	assertNoDefaultReviewFile(t, dir)
}

func TestAppend_UnknownType(t *testing.T) {
	dir := t.TempDir()
	err := Append(dir, Type("unknown"), validFinding("find-1"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown review type")
}

// --- UpdateStatus ---

func TestUpdateStatus_HappyPath(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Append(dir, TypeDefault, validFinding("find-1")))
	require.NoError(t, UpdateStatus(dir, TypeDefault, "find-1", "resolved"))
	findings, err := Load(dir, TypeDefault)
	require.NoError(t, err)
	assert.Equal(t, "resolved", findings[0].Status)
}

func TestUpdateStatus_UnknownID(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Append(dir, TypeDefault, validFinding("find-1")))
	err := UpdateStatus(dir, TypeDefault, "no-such-id", "resolved")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Append(dir, TypeDefault, validFinding("find-1")))
	err := UpdateStatus(dir, TypeDefault, "find-1", "done")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be")
}

func TestUpdateStatus_UnknownType(t *testing.T) {
	dir := t.TempDir()
	err := UpdateStatus(dir, Type("unknown"), "find-1", "resolved")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown review type")
}

func TestUpdateStatus_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	// File doesn't exist: Load returns empty slice, so ID will never be found.
	err := UpdateStatus(dir, TypeDefault, "find-1", "resolved")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateStatus_CorruptNeighbourBlocksUpdate(t *testing.T) {
	// A corrupt finding anywhere in the file blocks all UpdateStatus calls via Load's validation.
	// This is intentional: if the file is corrupt, we refuse to produce a partial update.
	dir := t.TempDir()
	require.NoError(t, Append(dir, TypeDefault, validFinding("find-1")))
	// Directly corrupt the file by appending an invalid finding.
	path := filepath.Join(dir, "review.yml")
	data, _ := os.ReadFile(path)
	data = append(data, []byte("- id: INVALID\n  feedback: corrupt\n  status: open\n")...)
	require.NoError(t, os.WriteFile(path, data, 0644))
	err := UpdateStatus(dir, TypeDefault, "find-1", "resolved")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid finding")
}

// --- kebab-case boundary ---

func TestValidate_SingleSegmentIDIsValid(t *testing.T) {
	// Single-segment lowercase IDs like "fix" are valid kebab-case by design.
	f := Finding{ID: "fix", Feedback: "feedback", Status: "open"}
	assert.NoError(t, validate(f))
}

// --- Append rejection does not write non-default types ---

func TestAppend_RejectsInvalidStatus_DocsType(t *testing.T) {
	dir := t.TempDir()
	f := Finding{ID: "find-1", Feedback: "feedback", Status: "todo"}
	err := Append(dir, TypeDocs, f)
	require.Error(t, err)
	_, statErr := os.Stat(filepath.Join(dir, "review-docs.yml"))
	assert.True(t, os.IsNotExist(statErr), "review-docs.yml must not exist after a rejected write")
}

// --- helpers ---

func assertNoDefaultReviewFile(t *testing.T, dir string) {
	t.Helper()
	_, err := os.Stat(filepath.Join(dir, "review.yml"))
	assert.True(t, os.IsNotExist(err), "review.yml must not exist after a rejected write")
}
