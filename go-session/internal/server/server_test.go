package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalHandler_MissingPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/action/terminal", nil)
	w := httptest.NewRecorder()
	server.TerminalHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "path parameter is required")
}

func TestTerminalHandler_NonExistentPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/action/terminal?path=/nonexistent/path/xyz123", nil)
	w := httptest.NewRecorder()
	server.TerminalHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTerminalHandler_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "testfile")
	require.NoError(t, err)
	f.Close()

	req := httptest.NewRequest(http.MethodGet, "/action/terminal?path="+filepath.ToSlash(f.Name()), nil)
	w := httptest.NewRecorder()
	server.TerminalHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFinderHandler_MissingPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/action/finder", nil)
	w := httptest.NewRecorder()
	server.FinderHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "path parameter is required")
}

func TestFinderHandler_NonExistentPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/action/finder?path=/nonexistent/path/xyz123", nil)
	w := httptest.NewRecorder()
	server.FinderHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFinderHandler_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "testfile")
	require.NoError(t, err)
	f.Close()

	req := httptest.NewRequest(http.MethodGet, "/action/finder?path="+filepath.ToSlash(f.Name()), nil)
	w := httptest.NewRecorder()
	server.FinderHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
