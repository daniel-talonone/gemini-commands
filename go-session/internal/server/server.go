package server

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"sort"

	"github.com/daniel-talonone/gemini-commands/internal/dashboard"
)

//go:embed template.html
var templateHTML string

// PageData is passed to the HTML template on every request.
type PageData struct {
	Features     []dashboard.FeatureState
	AllRepos     []string
	RepoFilter   string
	StatusFilter string
}

// Server is the dashboard HTTP server.
type Server struct {
	port int
	http *http.Server
}

// New creates a new Server listening on the given port.
func New(port int) *Server {
	return &Server{port: port}
}

// Start parses the template, registers routes, and begins listening.
// Blocks until the server stops. Returns nil on clean shutdown (ErrServerClosed).
func (s *Server) Start() error {
	tmpl, err := template.New("dashboard").Funcs(template.FuncMap{
		"safeURL": func(s string) template.URL { return template.URL(s) },
	}).Parse(templateHTML)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.makeHandler(tmpl))
	mux.HandleFunc("/action/terminal", TerminalHandler)

	addr := fmt.Sprintf(":%d", s.port)
	s.http = &http.Server{Addr: addr, Handler: mux}

	fmt.Printf("Dashboard running at http://localhost%s\n", addr)
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully drains in-flight requests then stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

// TerminalHandler handles GET /action/terminal?path=<dir> by opening a new
// Terminal.app window at the given directory. Returns 400 if path is missing
// or not an existing directory, 500 if the open command fails, 204 on success.
func TerminalHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter is required", http.StatusBadRequest)
		return
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		http.Error(w, "path is not an existing directory", http.StatusBadRequest)
		return
	}
	if err := exec.Command("open", "-a", "Terminal", path).Run(); err != nil {
		http.Error(w, "failed to open terminal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) makeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		features, err := dashboard.ScanAll()
		if err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		repoFilter := r.URL.Query().Get("repo")
		statusFilter := r.URL.Query().Get("status")

		// Reject unknown status filter values.
		if statusFilter != "" && statusFilter != "running" && statusFilter != "done" && statusFilter != "idle" {
			http.Error(w, "invalid status filter: use running, idle, or done", http.StatusBadRequest)
			return
		}

		// Collect unique repos for the filter dropdown.
		repoSet := map[string]struct{}{}
		for _, f := range features {
			if f.Repo != "" {
				repoSet[f.Repo] = struct{}{}
			}
		}
		allRepos := make([]string, 0, len(repoSet))
		for repo := range repoSet {
			allRepos = append(allRepos, repo)
		}
		sort.Strings(allRepos)

		// Apply filters.
		filtered := make([]dashboard.FeatureState, 0, len(features))
		for _, f := range features {
			if repoFilter != "" && f.Repo != repoFilter {
				continue
			}
			if statusFilter != "" {
				switch statusFilter {
				case "running":
					if !f.IsRunning {
						continue
					}
				case "done":
					if !f.AllDone {
						continue
					}
				case "idle":
					if f.IsRunning || f.AllDone {
						continue
					}
				}
			}
			filtered = append(filtered, f)
		}

		data := PageData{
			Features:     filtered,
			AllRepos:     allRepos,
			RepoFilter:   repoFilter,
			StatusFilter: statusFilter,
		}

		// Buffer template output — prevents partial responses on error.
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(buf.Bytes())
	}
}
