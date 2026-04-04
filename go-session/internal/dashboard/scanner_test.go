package dashboard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/dashboard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeFeatureDir(t *testing.T, root, org, repo, storyID string) string {
	t.Helper()
	dir := filepath.Join(root, org, repo, storyID)
	require.NoError(t, os.MkdirAll(dir, 0755))
	return dir
}

func TestScanRoot_EmptyDir(t *testing.T) {
	root := t.TempDir()
	results, err := dashboard.ScanRoot(root)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestScanRoot_MissingDir(t *testing.T) {
	results, err := dashboard.ScanRoot("/nonexistent/path/xyz/features")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestScanRoot_ReadsFeatureDir(t *testing.T) {
	root := t.TempDir()
	dir := makeFeatureDir(t, root, "myorg", "myrepo", "sc-1234")

	statusYAML := "mode: auto\nrepo: myorg/myrepo\nbranch: sc-1234\npid: 0\npipeline_step: plan\nstarted_at: ''\nupdated_at: ''\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "status.yaml"), []byte(statusYAML), 0644))

	planYAML := "- id: s1\n  status: done\n  tasks:\n    - id: t1\n      status: done\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte(planYAML), 0644))

	results, err := dashboard.ScanRoot(root)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "sc-1234", results[0].StoryID)
	assert.Equal(t, "myorg/myrepo", results[0].Repo)
	assert.True(t, results[0].AllDone)
	assert.Equal(t, "plan", results[0].PipelineStep)
}

func TestScanRoot_MissingStatusYAML(t *testing.T) {
	root := t.TempDir()
	makeFeatureDir(t, root, "myorg", "myrepo", "sc-1234")

	results, err := dashboard.ScanRoot(root)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.False(t, results[0].HasStatus)
	assert.False(t, results[0].IsRunning)
	assert.Equal(t, "myorg/myrepo", results[0].Repo)
}

func TestScanRoot_MultipleFeatures(t *testing.T) {
	root := t.TempDir()
	makeFeatureDir(t, root, "org1", "repo1", "sc-1")
	makeFeatureDir(t, root, "org2", "repo2", "sc-2")

	results, err := dashboard.ScanRoot(root)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestScanRoot_SkipsNonDirectories(t *testing.T) {
	root := t.TempDir()
	makeFeatureDir(t, root, "myorg", "myrepo", "sc-1234")
	// Place a stray file at the org level — should not be walked as a repo
	require.NoError(t, os.WriteFile(filepath.Join(root, "myorg", "stray.txt"), []byte("x"), 0644))

	results, err := dashboard.ScanRoot(root)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestScanRoot_PartialPlanYAML(t *testing.T) {
	root := t.TempDir()
	dir := makeFeatureDir(t, root, "myorg", "myrepo", "sc-1234")

	planYAML := "- id: s1\n  status: todo\n  tasks:\n    - id: t1\n      status: done\n    - id: t2\n      status: todo\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte(planYAML), 0644))

	results, err := dashboard.ScanRoot(root)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.False(t, results[0].AllDone)
	assert.Equal(t, "t1", results[0].LastDoneTask)
}
