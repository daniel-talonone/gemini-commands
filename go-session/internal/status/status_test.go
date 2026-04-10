package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// --- Write ---

func TestWrite_CreatesFileWhenAbsent(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Write(dir, "plan-done", "org/repo", "main"))

	// Write is a no-op when status.yaml does not exist
	_, err := os.Stat(filepath.Join(dir, "status.yaml"))
	assert.True(t, os.IsNotExist(err), "Write must not scaffold a new status.yaml")
}

func TestWrite_UpdatesExistingFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, "org/repo", "main", "/work", "", ""))

	require.NoError(t, Write(dir, "plan-done", "", ""))

	s, err := LoadStatus(dir)
	require.NoError(t, err)
	assert.Equal(t, "plan-done", s.PipelineStep)
	assert.Equal(t, "org/repo", s.Repo, "repo must be preserved when Write is called with empty repo")
}

func TestWrite_PreservesTimestamps(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, "org/repo", "main", "/work", "", ""))

	before, err := LoadStatus(dir)
	require.NoError(t, err)

	require.NoError(t, Write(dir, "step", "", ""))

	after, err := LoadStatus(dir)
	require.NoError(t, err)
	assert.Equal(t, before.StartedAt, after.StartedAt, "started_at must not change on Write")
}

// --- Create ---

func TestCreate_WritesAllFields(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Create(dir, "org/repo", "feat", "/work", "https://example.com/sc-1", "auto"))

	s, err := LoadStatus(dir)
	require.NoError(t, err)

	assert.Equal(t, "org/repo", s.Repo)
	assert.Equal(t, "feat", s.Branch)
	assert.Equal(t, "/work", s.WorkDir)
	assert.Equal(t, "https://example.com/sc-1", s.StoryURL)
	assert.Equal(t, "auto", s.Mode)
	assert.Equal(t, "new", s.PipelineStep)
	assert.Equal(t, 0, s.PID)
	assert.Equal(t, "", s.ClonePath, "clone_path must be present as empty string")
	assert.Equal(t, "", s.Error, "error must be present as empty string")

	_, err = time.Parse(time.RFC3339, s.StartedAt)
	require.NoError(t, err, "started_at must be a valid RFC3339 timestamp")
	_, err = time.Parse(time.RFC3339, s.UpdatedAt)
	require.NoError(t, err, "updated_at must be a valid RFC3339 timestamp")
}

func TestCreate_EmptyFieldsSerializedAsEmptyStrings(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, "", "", "", "", ""))

	raw, err := os.ReadFile(filepath.Join(dir, "status.yaml"))
	require.NoError(t, err)

	content := string(raw)
	assert.Contains(t, content, "clone_path:", "clone_path key must always be present in YAML")
	assert.Contains(t, content, "error:", "error key must always be present in YAML")
	assert.Contains(t, content, "story_url:", "story_url key must always be present in YAML")
}

func TestCreate_Idempotent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, "org/repo", "main", "/work", "", ""))

	// Modify the file to simulate live runtime state
	s, err := LoadStatus(dir)
	require.NoError(t, err)
	s.PipelineStep = "implement-done"
	data, err := yaml.Marshal(s)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "status.yaml"), data, 0644))

	// Second Create call must be a no-op
	require.NoError(t, Create(dir, "other/repo", "other-branch", "/other", "", ""))

	after, err := LoadStatus(dir)
	require.NoError(t, err)
	assert.Equal(t, "implement-done", after.PipelineStep, "Create must not overwrite existing status.yaml")
	assert.Equal(t, "org/repo", after.Repo, "Create must not overwrite existing status.yaml")
}

func TestCreate_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Create(dir, "org/repo", "main", "/work", "", ""))

	// Temp file must not linger after a successful Create
	_, err := os.Stat(filepath.Join(dir, "status.yaml.tmp"))
	assert.True(t, os.IsNotExist(err), "status.yaml.tmp must be cleaned up after Create")
}
