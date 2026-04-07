package git

import (
	"os/exec"
	"strings"
)

// RemoteURL returns the git remote origin URL, or "" if not in a git repo.
// exec.Command is intentionally kept in the CLI layer, not in internal/commands.
func RemoteURL() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// OrgRepo returns "org/repo" derived from git remote origin, or "" if unavailable.
func OrgRepo() string {
	remoteURL := RemoteURL()
	if remoteURL == "" {
		return ""
	}
	return ParseOrgRepo(remoteURL)
}

// WorkDir returns the repo root path from git rev-parse --show-toplevel, or "" if unavailable.
func WorkDir() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// CurrentBranch returns the current git branch name, or "" if unavailable.
func CurrentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" { // detached HEAD state
		return ""
	}
	return branch
}

// ParseOrgRepo extracts "org/repo" from SSH and HTTPS git remote URLs.
// SSH:   git@github.com:org/repo.git  → org/repo
// HTTPS: https://github.com/org/repo.git → org/repo
func ParseOrgRepo(remoteURL string) string {
	url := strings.TrimSuffix(remoteURL, ".git")
	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			return parts[1]
		}
		return ""
	}
	idx := strings.Index(url, "://")
	if idx == -1 {
		return ""
	}
	rest := url[idx+3:]
	slash := strings.Index(rest, "/")
	if slash == -1 {
		return ""
	}
	return rest[slash+1:]
}
