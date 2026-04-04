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

func TestLoadContext_OutputsFilesAsXML(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "description.md"), []byte("hello"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte("world"), 0644))

	out, err := commands.LoadContext(dir)
	require.NoError(t, err)
	assert.Contains(t, out, "<file name=\"description.md\">\nhello\n</file>")
	assert.Contains(t, out, "<file name=\"plan.yml\">\nworld\n</file>")
}

func TestLoadContext_SortedAlphabetically(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "z.md"), []byte("last"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte("first"), 0644))

	out, err := commands.LoadContext(dir)
	require.NoError(t, err)
	aIdx := strings.Index(out, "name=\"a.md\"")
	zIdx := strings.Index(out, "name=\"z.md\"")
	assert.Greater(t, zIdx, aIdx, "a.md block should appear before z.md block")
}

func TestLoadContext_ExcludesUnderscoreFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "_SUMMARY.md"), []byte("generated"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "description.md"), []byte("real"), 0644))

	out, err := commands.LoadContext(dir)
	require.NoError(t, err)
	assert.NotContains(t, out, "_SUMMARY.md")
	assert.Contains(t, out, "description.md")
}

func TestLoadContext_IncludesOnlyTargetExtensions(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore me"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte("include me"), 0644))

	out, err := commands.LoadContext(dir)
	require.NoError(t, err)
	assert.Contains(t, out, "plan.yml")
	assert.NotContains(t, out, "notes.txt")
}

func TestLoadContext_SupportedExtensions(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte("md"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.yml"), []byte("yml"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "c.yaml"), []byte("yaml"), 0644))

	out, err := commands.LoadContext(dir)
	require.NoError(t, err)
	assert.Contains(t, out, "name=\"a.md\"")
	assert.Contains(t, out, "name=\"b.yml\"")
	assert.Contains(t, out, "name=\"c.yaml\"")
}

func TestLoadContext_ErrorOnMissingFeatureDir(t *testing.T) {
	_, err := commands.LoadContext("/nonexistent/path/xyz")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestLoadContext_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	out, err := commands.LoadContext(dir)
	require.NoError(t, err)
	assert.Equal(t, "", out)
}

func TestLoadContext_FileSeparatedByBlankLine(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte("first"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.md"), []byte("second"), 0644))

	out, err := commands.LoadContext(dir)
	require.NoError(t, err)
	assert.Contains(t, out, "</file>\n\n<file")
}
