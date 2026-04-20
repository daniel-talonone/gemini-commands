package plan_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	archPath := filepath.Join(dir, "architecture.md")
	require.NoError(t, os.WriteFile(archPath, []byte("Test Architecture"), 0644))

	arch, err := plan.LoadArchitecture(dir)
	require.NoError(t, err)
	assert.Equal(t, "Test Architecture", arch)
}

func TestLoad_NotFound(t *testing.T) {
	dir := t.TempDir()
	arch, err := plan.LoadArchitecture(dir)
	require.NoError(t, err)
	assert.Equal(t, "", arch)
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	content := "My awesome architecture"
	err := plan.WriteArchitecture(dir, content)
	require.NoError(t, err)

	readContent, err := os.ReadFile(filepath.Join(dir, "architecture.md"))
	require.NoError(t, err)
	assert.Equal(t, content, string(readContent))
}

func TestWrite_NoDir(t *testing.T) {
	err := plan.WriteArchitecture("/tmp/nonexistent-dir-for-testing", "content")
	require.Error(t, err)
}
