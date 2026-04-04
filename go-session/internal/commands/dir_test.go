package commands_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFeature_CreatesAllFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sc-1234")
	err := commands.CreateFeature(target, "", "")
	require.NoError(t, err)
	for _, name := range []string{"plan.yml", "questions.yml", "review.yml", "log.md", "pr.md", "status.yaml"} {
		_, statErr := os.Stat(filepath.Join(target, name))
		assert.NoError(t, statErr, "expected %s to exist", name)
	}
}

func TestCreateFeature_PlaceholderContent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, commands.CreateFeature(dir, "", ""))
	for _, name := range []string{"plan.yml", "questions.yml", "review.yml"} {
		content, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err)
		assert.Equal(t, "[]\n", string(content), "%s should contain []", name)
	}
	log, _ := os.ReadFile(filepath.Join(dir, "log.md"))
	assert.Contains(t, string(log), "# Work Log")
	pr, _ := os.ReadFile(filepath.Join(dir, "pr.md"))
	assert.Contains(t, string(pr), "# Pull Request")
}

func TestCreateFeature_StatusYAMLContent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, commands.CreateFeature(dir, "", ""))
	content, err := os.ReadFile(filepath.Join(dir, "status.yaml"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "pid: 0")
	assert.Contains(t, s, "pipeline_step:")
}

func TestCreateFeature_StatusYAMLPopulatesRepoAndBranch(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, commands.CreateFeature(dir, "myorg/myrepo", "sc-1234"))
	content, err := os.ReadFile(filepath.Join(dir, "status.yaml"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "repo: myorg/myrepo")
	assert.Contains(t, s, "branch: sc-1234")
	assert.NotContains(t, s, "started_at: ''")
	assert.Contains(t, s, "started_at: '20")
	assert.NotContains(t, s, "updated_at: ''")
	assert.Contains(t, s, "updated_at: '20")
}

func TestCreateFeature_StatusYAMLEmptyRepoAndBranch(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, commands.CreateFeature(dir, "", ""))
	content, err := os.ReadFile(filepath.Join(dir, "status.yaml"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "repo: ''")
	assert.Contains(t, s, "branch: ''")
}

func TestCreateFeature_Idempotent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, commands.CreateFeature(dir, "", ""))
	assert.NoError(t, commands.CreateFeature(dir, "", ""), "second call should not error")
}

func TestCreateFeature_DoesNotOverwriteExistingFiles(t *testing.T) {
	dir := t.TempDir()
	// Write a live status.yaml with real data before calling CreateFeature
	statusPath := filepath.Join(dir, "status.yaml")
	liveContent := "mode: auto\nrepo: org/repo\nbranch: sc-1\npid: 12345\npipeline_step: implement\nstarted_at: '2026-01-01T00:00:00Z'\nupdated_at: '2026-01-01T00:00:00Z'\n"
	require.NoError(t, os.WriteFile(statusPath, []byte(liveContent), 0644))

	require.NoError(t, commands.CreateFeature(dir, "other/repo", "other-branch"))

	content, err := os.ReadFile(statusPath)
	require.NoError(t, err)
	assert.Equal(t, liveContent, string(content), "CreateFeature must not overwrite existing status.yaml")
}

func TestResolveFeatureDir_ExplicitAbsPath(t *testing.T) {
	dir := t.TempDir()
	result, err := commands.ResolveFeatureDir("/absolute/path/sc-1", dir, "")
	require.NoError(t, err)
	assert.Equal(t, "/absolute/path/sc-1", result)
}

func TestResolveFeatureDir_DotPath(t *testing.T) {
	dir := t.TempDir()
	result, err := commands.ResolveFeatureDir(".features/sc-1", dir, "")
	require.NoError(t, err)
	assert.Equal(t, ".features/sc-1", result)
}

func TestResolveFeatureDir_LocalFeaturesDir(t *testing.T) {
	cwd := t.TempDir()
	local := filepath.Join(cwd, ".features", "sc-1234")
	require.NoError(t, os.MkdirAll(local, 0755))
	result, err := commands.ResolveFeatureDir("sc-1234", cwd, "")
	require.NoError(t, err)
	assert.Equal(t, local, result)
}

func TestResolveFeatureDir_HTTPSRemote(t *testing.T) {
	dir := t.TempDir()
	result, err := commands.ResolveFeatureDir(
		"sc-1234", dir, "https://github.com/myorg/myrepo.git",
	)
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(result, "myorg/myrepo/sc-1234"))
	assert.Contains(t, result, ".ai-session/features/")
}

func TestResolveFeatureDir_SSHRemote(t *testing.T) {
	dir := t.TempDir()
	result, err := commands.ResolveFeatureDir(
		"sc-1234", dir, "git@github.com:myorg/myrepo.git",
	)
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(result, "myorg/myrepo/sc-1234"))
}

func TestResolveFeatureDir_NoRemoteNoLocal(t *testing.T) {
	dir := t.TempDir()
	_, err := commands.ResolveFeatureDir("sc-1234", dir, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sc-1234")
}
