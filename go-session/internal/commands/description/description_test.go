package description_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/commands/description"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDescription(t *testing.T) {
	dir := t.TempDir()
	descPath := filepath.Join(dir, "description.md")
	require.NoError(t, os.WriteFile(descPath, []byte("Test Description"), 0644))

	desc, err := description.LoadDescription(dir)
	require.NoError(t, err)
	assert.Equal(t, "Test Description", desc)
}

func TestLoadDescription_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := description.LoadDescription(dir)
	assert.Error(t, err)
}

func TestLoadArchitecture(t *testing.T) {
	dir := t.TempDir()
	archPath := filepath.Join(dir, "architecture.md")
	require.NoError(t, os.WriteFile(archPath, []byte("Test Architecture"), 0644))

	arch, err := description.LoadArchitecture(dir)
	require.NoError(t, err)
	assert.Equal(t, "Test Architecture", arch)
}

func TestLoadArchitecture_NotFound(t *testing.T) {
	dir := t.TempDir()
	arch, err := description.LoadArchitecture(dir)
	require.NoError(t, err)
	assert.Equal(t, "", arch)
}
