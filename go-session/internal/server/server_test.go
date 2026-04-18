package server_test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daniel-talonone/gemini-commands/internal/dashboard"
	"github.com/daniel-talonone/gemini-commands/internal/feature"
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
	require.NoError(t, f.Close())

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
	require.NoError(t, f.Close())

	req := httptest.NewRequest(http.MethodGet, "/action/finder?path="+filepath.ToSlash(f.Name()), nil)
	w := httptest.NewRecorder()
	server.FinderHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// mockScanner implements the Scanner interface for testing purposes.
type mockScanner struct {
	features []dashboard.FeatureState
	err      error
}

func (m *mockScanner) ScanAll() ([]dashboard.FeatureState, error) {
	return m.features, m.err
}

// storyIDs extracts StoryID from a slice of FeatureState for easy assertion.
func storyIDs(features []dashboard.FeatureState) []string {
	ids := make([]string, len(features))
	for i, f := range features {
		ids[i] = f.StoryID
	}
	return ids
}

// parsedOrder parses space-separated StoryIDs from the dummy template output.
func parsedOrder(body string) []string {
	return strings.Fields(strings.TrimSpace(body))
}

func TestSortingLogic(t *testing.T) {
	mockData := []dashboard.FeatureState{
		{
			StoryID:   "feature-c",
			Repo:      "repo-a",
			UpdatedAt: time.Date(2023, time.January, 3, 0, 0, 0, 0, time.UTC),
			StartedAt: time.Date(2023, time.January, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			StoryID:   "feature-a",
			Repo:      "repo-b",
			IsRunning: true,
			UpdatedAt: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			StartedAt: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			StoryID:   "feature-b",
			Repo:      "repo-a",
			UpdatedAt: time.Date(2023, time.January, 2, 0, 0, 0, 0, time.UTC),
			StartedAt: time.Date(2023, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
	}
	mockS := &mockScanner{features: mockData}

	// Dummy template outputs StoryIDs space-separated for easy order verification.
	tmpl, err := template.New("test").Parse(`{{range .Features}}{{.StoryID}} {{end}}`)
	require.NoError(t, err)

	srv := server.New(8080, mockS)

	tests := []struct {
		name          string
		url           string
		expectedOrder []string
	}{
		{
			name:          "Default sort is updated desc",
			url:           "/",
			expectedOrder: []string{"feature-c", "feature-b", "feature-a"},
		},
		{
			name:          "Sort by updated asc",
			url:           "/?sort=updated&order=asc",
			expectedOrder: []string{"feature-a", "feature-b", "feature-c"},
		},
		{
			name:          "Sort by updated desc",
			url:           "/?sort=updated&order=desc",
			expectedOrder: []string{"feature-c", "feature-b", "feature-a"},
		},
		{
			name:          "Sort by started asc",
			url:           "/?sort=started&order=asc",
			expectedOrder: []string{"feature-a", "feature-b", "feature-c"},
		},
		{
			name:          "Sort by started desc",
			url:           "/?sort=started&order=desc",
			expectedOrder: []string{"feature-c", "feature-b", "feature-a"},
		},
		{
			name:          "Unknown sort key falls back to updated field with given order",
			url:           "/?sort=unknown&order=asc",
			expectedOrder: []string{"feature-a", "feature-b", "feature-c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()
			srv.MakeListHandler(tmpl).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tt.expectedOrder, parsedOrder(rr.Body.String()))
		})
	}
}

func TestSortingWithFilters(t *testing.T) {
	mockData := []dashboard.FeatureState{
		{StoryID: "feature-c", Repo: "repo-a", AllDone: true, UpdatedAt: time.Date(2023, time.January, 3, 0, 0, 0, 0, time.UTC)},
		{StoryID: "feature-a", Repo: "repo-b", IsRunning: true, UpdatedAt: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{StoryID: "feature-b", Repo: "repo-a", UpdatedAt: time.Date(2023, time.January, 2, 0, 0, 0, 0, time.UTC)},
	}
	mockS := &mockScanner{features: mockData}

	tmpl, err := template.New("test").Parse(`{{range .Features}}{{.StoryID}} {{end}}`)
	require.NoError(t, err)

	srv := server.New(8080, mockS)

	tests := []struct {
		name          string
		url           string
		expectedOrder []string
	}{
		{
			name:          "Sort by updated asc with repo filter",
			url:           "/?repo=repo-a&sort=updated&order=asc",
			expectedOrder: []string{"feature-b", "feature-c"},
		},
		{
			name:          "Sort by updated desc with done status filter",
			url:           "/?status=done&sort=updated&order=desc",
			expectedOrder: []string{"feature-c"},
		},
		{
			name:          "Sort by updated desc with running status filter",
			url:           "/?status=running&sort=updated&order=desc",
			expectedOrder: []string{"feature-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()
			srv.MakeListHandler(tmpl).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tt.expectedOrder, parsedOrder(rr.Body.String()))
		})
	}
}

func TestSortFeaturesByUpdatedAt(t *testing.T) {
	features := []dashboard.FeatureState{
		{StoryID: "a", UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{StoryID: "b", UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)},
		{StoryID: "c", UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
	}

	cp := make([]dashboard.FeatureState, len(features))

	copy(cp, features)
	server.SortFeatures(cp, "updated", "asc")
	assert.Equal(t, []string{"a", "c", "b"}, storyIDs(cp))

	copy(cp, features)
	server.SortFeatures(cp, "updated", "desc")
	assert.Equal(t, []string{"b", "c", "a"}, storyIDs(cp))
}

func TestSortFeaturesByStartedAt(t *testing.T) {
	features := []dashboard.FeatureState{
		{StoryID: "a", StartedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)},
		{StoryID: "b", StartedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{StoryID: "c", StartedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
	}

	cp := make([]dashboard.FeatureState, len(features))

	copy(cp, features)
	server.SortFeatures(cp, "started", "asc")
	assert.Equal(t, []string{"b", "c", "a"}, storyIDs(cp))

	copy(cp, features)
	server.SortFeatures(cp, "started", "desc")
	assert.Equal(t, []string{"a", "c", "b"}, storyIDs(cp))
}

func TestSortFeaturesZeroTimesLast(t *testing.T) {
	features := []dashboard.FeatureState{
		{StoryID: "a", UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{StoryID: "b"}, // zero UpdatedAt
		{StoryID: "c", UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
	}

	cp := make([]dashboard.FeatureState, len(features))

	copy(cp, features)
	server.SortFeatures(cp, "updated", "desc")
	assert.Equal(t, []string{"c", "a", "b"}, storyIDs(cp), "zero times should sort last in desc")

	copy(cp, features)
	server.SortFeatures(cp, "updated", "asc")
	assert.Equal(t, []string{"b", "a", "c"}, storyIDs(cp), "zero times should sort first in asc")
}

// TestHandlerWithRealTemplate renders the actual template.html to catch
// field name mismatches between PageData/FeatureState and the template.
func TestHandlerWithRealTemplate(t *testing.T) {
	tmplContent, err := os.ReadFile("template.html")
	require.NoError(t, err)

	funcMap := template.FuncMap{
		"safeURL": func(s string) template.URL { return template.URL(s) },
	}
	tmpl, err := template.New("dashboard").Funcs(funcMap).Parse(string(tmplContent))
	require.NoError(t, err)

	now := time.Now()
	mockS := &mockScanner{features: []dashboard.FeatureState{
		{
			StoryID:            "sc-1",
			Repo:               "org/repo",
			StartedAt:          now.Add(-1 * time.Hour),
			UpdatedAt:          now,
			FormattedStartedAt: "1 hour ago",
			FormattedUpdatedAt: "just now",
		},
	}}
	srv := server.New(8080, mockS)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.MakeListHandler(tmpl).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "real template should render without errors")
	assert.Contains(t, rr.Body.String(), "sc-1")
	assert.Contains(t, rr.Body.String(), "1 hour ago")
	assert.Contains(t, rr.Body.String(), "just now")
}

func TestFeatureDetailHandler(t *testing.T) {
	// This template is a simplified version of the real one, only containing
	// the parts needed for this test.
	tmplContent := `{{define "feature_detail"}}<h1>Feature: {{.ID}}</h1><a href="/">Back</a>{{end}}`
	tmpl, err := template.New("dashboard").Parse(tmplContent)
	require.NoError(t, err)

	mockS := &mockScanner{} // No features needed for this test
	srv := server.New(8080, mockS)

	t.Run("feature not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/feature/non-existent-feature", nil)
		rr := httptest.NewRecorder()
		srv.MakeFeatureDetailHandler(tmpl).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("feature found", func(t *testing.T) {
		featureID := "sc-12345"
		repo := "org/repo_name"

		home, err := os.UserHomeDir()
		require.NoError(t, err)
		featureDir := filepath.Join(home, ".features", repo, featureID)
		require.NoError(t, feature.CreateFeature(featureDir, repo, "main", ""))
		require.NoError(t, os.WriteFile(filepath.Join(featureDir, "description.md"), []byte("# "+featureID), 0755))
		defer func() { _ = os.RemoveAll(featureDir) }()

		mockS.features = []dashboard.FeatureState{{StoryID: featureID, Repo: repo}}

		req := httptest.NewRequest(http.MethodGet, "/feature/"+featureID, nil)
		rr := httptest.NewRecorder()
		srv.MakeFeatureDetailHandler(tmpl).ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "<h1>Feature: sc-12345</h1>")
		assert.Contains(t, rr.Body.String(), `<a href="/">Back</a>`)
	})
}
