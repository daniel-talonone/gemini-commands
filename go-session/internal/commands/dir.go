package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daniel-talonone/gemini-commands/internal/git"
)

// CreateFeature creates a feature directory with placeholder files.
// repo, branch, and workDir are written into status.yaml if provided; pass "" to leave them empty.
// Idempotent: succeeds if the directory already exists and never overwrites existing files.
func CreateFeature(featureDir, repo, branch, workDir string) error {
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		return fmt.Errorf("creating feature directory: %w", err)
	}

	repoVal, branchVal, workDirVal := "''", "''", "''"
	if repo != "" {
		repoVal = repo
	}
	if branch != "" {
		branchVal = branch
	}
	if workDir != "" {
		workDirVal = workDir
	}
	now := time.Now().Format(time.RFC3339)

	files := map[string]string{
		"plan.yml":      "[]\n",
		"questions.yml": "[]\n",
		"review.yml":    "[]\n",
		"log.md":        "# Work Log\n*(This section is intentionally left blank.)*\n",
		"pr.md":         "# Pull Request\n*(This section is intentionally left blank.)*\n",
		"status.yaml":   fmt.Sprintf("mode: ''\nrepo: %s\nbranch: %s\nwork_dir: %s\npid: 0\npipeline_step: ''\nstarted_at: '%s'\nupdated_at: '%s'\n", repoVal, branchVal, workDirVal, now, now),
	}

	for name, content := range files {
		path := filepath.Join(featureDir, name)
		if _, err := os.Stat(path); err == nil {
			continue // file already exists — never overwrite live data
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}
	return nil
}

// ResolveFeatureDir resolves the full path to a feature directory.
// cwd is used for the local .features/ backward-compat check.
// remoteURL is the git remote origin URL (pass "" if not in a git repo).
func ResolveFeatureDir(storyID, cwd, remoteURL string) (string, error) {
	if strings.Contains(storyID, "/") ||
		strings.HasPrefix(storyID, ".") ||
		strings.HasPrefix(storyID, "~") {
		return storyID, nil
	}

	localDir := filepath.Join(cwd, ".features", storyID)
	if info, err := os.Stat(localDir); err == nil && info.IsDir() {
		return localDir, nil
	}

	if remoteURL == "" {
		return "", fmt.Errorf(
			"cannot resolve %q: not an explicit path, no local .features/%s directory found, "+
				"and no git remote URL provided — run from inside a git repository or pass a full path",
			storyID, storyID,
		)
	}

	orgRepo := git.ParseOrgRepo(remoteURL)
	if orgRepo == "" {
		return "", fmt.Errorf("cannot parse org/repo from remote URL: %s", remoteURL)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".ai-session", "features", orgRepo, storyID), nil
}
