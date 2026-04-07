package log_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLogFile_CreatesFileWithHeader(t *testing.T) {
	dir := t.TempDir()
	err := log.CreateLogFile(dir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Work Log")
}

func TestCreateLogFile_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "log.md")
	initialContent := "existing content"
	require.NoError(t, os.WriteFile(logPath, []byte(initialContent), 0644))

	err := log.CreateLogFile(dir)
	require.NoError(t, err)

	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Equal(t, initialContent, string(content))
}

func TestAppendLog_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	err := log.AppendLog(dir, "hello world")
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(dir, "log.md"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "# Work Log")
	assert.Contains(t, s, "## [")
	assert.Contains(t, s, "hello world")
}

func TestAppendLog_AppendsSeparator(t *testing.T) {
	dir := t.TempDir()
	err := log.AppendLog(dir, "first")
	require.NoError(t, err)
	err = log.AppendLog(dir, "second")
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(dir, "log.md"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "# Work Log")
	assert.Equal(t, 2, strings.Count(s, "## ["), "should have two headers")
	assert.Contains(t, s, `first

## [`)
}

func TestAppendLog_TimestampFormat(t *testing.T) {
	dir := t.TempDir()
	err := log.AppendLog(dir, "msg")
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(dir, "log.md"))
	require.NoError(t, err)
	assert.Regexp(t, `## \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z\]`, string(content))
}

func TestAppendLog_ErrorOnMissingDir(t *testing.T) {
	err := log.AppendLog("/nonexistent/path/xyz", "msg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
