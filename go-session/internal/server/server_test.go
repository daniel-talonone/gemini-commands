package server_test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	// Full feature_detail template block to test markdown rendering
	tmplContent := `{{define "feature_detail"}}
<a href="/">← Back</a>
<h1>{{.ID}}</h1>
<p>
  {{if .Repo}}<strong>Repo:</strong> <a href="https://github.com/{{.Repo}}" target="_blank">{{.Repo}}</a>{{end}}
  {{if .Branch}} | <strong>Branch:</strong> {{.Branch}}{{end}}
  {{if .StoryURL}} | <a href="{{.StoryURL}}" target="_blank">Story</a>{{end}}
  {{if .PRURL}} | <a href="{{.PRURL}}" target="_blank">Pull Request</a>{{end}}
</p>
{{if .WorkDir}}
<p>
  <strong>Quick Launch:</strong>
  <a href="/action/finder?path={{.WorkDir | urlquery}}" title="Open in Finder">📁</a>
  <a href="{{printf "vscode://file%s" .WorkDir | safeURL}}" title="Open in VSCode">VSCode</a>
  <a href="/action/terminal?path={{.WorkDir | urlquery}}" title="Open Terminal">⬛</a>
</p>
{{end}}
{{if .Description}}
<details open>
  <summary>Description</summary>
  <div class="description-content">
    {{.Description}}
  </div>
</details>
{{end}}
{{end}}`

	funcMap := template.FuncMap{
		"safeURL": func(s string) template.URL { return template.URL(s) },
		"urlquery": func(s string) string { return url.QueryEscape(s) },
	}
	tmpl, err := template.New("dashboard").Funcs(funcMap).Parse(tmplContent)
	require.NoError(t, err)

	mockS := &mockScanner{} // No features needed for this test
	srv := server.New(8080, mockS)

	t.Run("feature not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/feature/non-existent-feature", nil)
		rr := httptest.NewRecorder()
		srv.MakeFeatureDetailHandler(tmpl).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("feature found with markdown description", func(t *testing.T) {
		featureID := "sc-12345"
		repo := "org/repo_name"

		home, err := os.UserHomeDir()
		require.NoError(t, err)
		featureDir := filepath.Join(home, ".features", repo, featureID)
		require.NoError(t, feature.CreateFeature(featureDir, repo, "main", ""))

		// Create description.md with diverse markdown content
		descriptionContent := "# Test Feature\n\n" +
			"**Bold text** and _italic text_\n\n" +
			"- Item 1\n- Item 2\n- Item 3\n\n" +
			"1. First item\n2. Second item\n\n" +
			"`code snippet`\n\n" +
			"```go\nfunc main() {\n  fmt.Println(\"test\")\n}\n```"

		require.NoError(t, os.WriteFile(filepath.Join(featureDir, "description.md"), []byte(descriptionContent), 0755))
		defer func() { _ = os.RemoveAll(featureDir) }()

		mockS.features = []dashboard.FeatureState{{StoryID: featureID, Repo: repo}}

		req := httptest.NewRequest(http.MethodGet, "/feature/"+featureID, nil)
		rr := httptest.NewRecorder()
		srv.MakeFeatureDetailHandler(tmpl).ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		body := rr.Body.String()
		// Verify feature metadata is present
		assert.Contains(t, body, "<h1>sc-12345</h1>")
		assert.Contains(t, body, "org/repo_name")

		// Verify markdown rendering: headings
		assert.Contains(t, body, "<h1>Test Feature</h1>", "expected h1 heading to be rendered")
		// Verify markdown rendering: bold and italic
		assert.Contains(t, body, "<strong>Bold text</strong>", "expected strong tag for bold text")
		assert.Contains(t, body, "<em>italic text</em>", "expected em tag for italic text")
		// Verify markdown rendering: lists
		assert.Contains(t, body, "<li>Item 1</li>", "expected list items to be rendered")
		assert.Contains(t, body, "<ol>", "expected ordered list to be rendered")
		// Verify markdown rendering: inline code
		assert.Contains(t, body, "<code>code snippet</code>", "expected inline code to be rendered")
		// Verify markdown rendering: code block
		assert.Contains(t, body, "<pre>", "expected pre tag for code block")
		// Verify description section structure
		assert.Contains(t, body, "<details open>", "expected details section to be open")
		assert.Contains(t, body, "<summary>Description</summary>", "expected description summary")
	})

	t.Run("missing description hides section", func(t *testing.T) {
		featureID := "sc-67890"
		repo := "org/repo_name"

		home, err := os.UserHomeDir()
		require.NoError(t, err)
		featureDir := filepath.Join(home, ".features", repo, featureID)
		require.NoError(t, feature.CreateFeature(featureDir, repo, "main", ""))
		// Do NOT create description.md
		defer func() { _ = os.RemoveAll(featureDir) }()

		mockS.features = []dashboard.FeatureState{{StoryID: featureID, Repo: repo}}

		req := httptest.NewRequest(http.MethodGet, "/feature/"+featureID, nil)
		rr := httptest.NewRecorder()
		srv.MakeFeatureDetailHandler(tmpl).ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		body := rr.Body.String()
		// Verify feature metadata is still rendered
		assert.Contains(t, body, "<h1>sc-67890</h1>", "expected feature ID to be rendered")
		assert.Contains(t, body, "org/repo_name", "expected repo to be rendered")
		// Verify description section is hidden
		assert.NotContains(t, body, "<details>", "expected no details section when description is missing")
		assert.NotContains(t, body, "<summary>Description</summary>", "expected no description summary")
	})
}
