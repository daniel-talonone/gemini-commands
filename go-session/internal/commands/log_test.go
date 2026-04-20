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

func TestAppendLog_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	err := commands.AppendLog(dir, "hello world")
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(dir, "log.md"))
	require.NoError(t, err)
	s := string(content)
	assert.True(t, strings.HasPrefix(s, "## ["), "should start with ## [")
	assert.Contains(t, s, "hello world")
}

func TestAppendLog_AppendsSeparator(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, commands.AppendLog(dir, "first"))
	require.NoError(t, commands.AppendLog(dir, "second"))
	content, err := os.ReadFile(filepath.Join(dir, "log.md"))
	require.NoError(t, err)
	s := string(content)
	assert.Equal(t, 2, strings.Count(s, "## ["), "should have two headers")
	assert.Contains(t, s, "first\n\n## [")
}

func TestAppendLog_TimestampFormat(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, commands.AppendLog(dir, "msg"))
	content, _ := os.ReadFile(filepath.Join(dir, "log.md"))
	s := string(content)
	assert.Contains(t, s, "T")
	assert.Contains(t, s, "Z]")
}

func TestAppendLog_ErrorOnMissingDir(t *testing.T) {
	err := commands.AppendLog("/nonexistent/path/xyz", "msg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
