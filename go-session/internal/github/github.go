package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var execCommand = exec.Command

type PRReviewThread struct {
	Comments []struct {
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Body string `json:"body"`
		Path string `json:"path"`
		Line int    `json:"line"`
	} `json:"comments"`
	IsResolved bool `json:"isResolved"`
}

type PRView struct {
	ReviewThreads []PRReviewThread `json:"reviewThreads"`
}

func GetUnresolvedReviewThreads(workDir, branch string) (string, error) {
	cmd := execCommand("gh", "pr", "view", "--json", "reviewThreads")
	cmd.Dir = workDir
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("gh command failed: %s: %w", stderr.String(), err)
	}

	var prView PRView
	if err := json.Unmarshal(out.Bytes(), &prView); err != nil {
		return "", fmt.Errorf("parsing gh output: %w", err)
	}

	var formattedThreads strings.Builder
	for _, thread := range prView.ReviewThreads {
		if !thread.IsResolved {
			for _, comment := range thread.Comments {
				var fileLine string
				if comment.Line == 0 {
					fileLine = comment.Path
				} else {
					fileLine = fmt.Sprintf("%s:%d", comment.Path, comment.Line)
				}
				fmt.Fprintf(&formattedThreads, `File: %s
Author: %s
Comment: %s

`,
					fileLine,
					comment.Author.Login,
					comment.Body)
			}
		}
	}

	return formattedThreads.String(), nil
}

// CreatePR creates a GitHub pull request and returns its URL.
// It first checks if a PR already exists for the given branch and returns an error if so.
func CreatePR(workDir, base, head, title, body string) (string, error) {
	cmdView := execCommand("gh", "pr", "view", head, "--json", "url")
	cmdView.Dir = workDir
	var outView bytes.Buffer
	var errView bytes.Buffer
	cmdView.Stdout = &outView
	cmdView.Stderr = &errView
	err := cmdView.Run()

	if err == nil {
		// PR exists, parse the URL and return an error
		var prInfo struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(outView.Bytes(), &prInfo); err != nil {
			return "", fmt.Errorf("parsing existing PR info: %w", err)
		}
		return "", fmt.Errorf("pr already exists for branch %s", head)
	}

	// If `gh pr view` returns an error, check if it's because no PR was found.
	// Otherwise, it's an actual error.
	if !strings.Contains(errView.String(), "no pull requests found") {
		return "", fmt.Errorf("failed to check for existing PR: %s: %w", errView.String(), err)
	}

	// PR does not exist, create it.
	// Write the body to a temp file to avoid shell-length limits with multiline content.
	args := []string{"pr", "create", "--base", base, "--head", head, "--title", title}
	var tmpBodyFile string
	if body != "" {
		f, err := os.CreateTemp("", "pr-body-*.md")
		if err != nil {
			return "", fmt.Errorf("creating body temp file: %w", err)
		}
		tmpBodyFile = f.Name()
		defer os.Remove(tmpBodyFile) //nolint:errcheck
		if _, err := f.WriteString(body); err != nil {
			f.Close() //nolint:errcheck
			return "", fmt.Errorf("writing body temp file: %w", err)
		}
		if err := f.Close(); err != nil {
			return "", fmt.Errorf("closing body temp file: %w", err)
		}
		args = append(args, "--body-file", tmpBodyFile)
	}

	cmdCreate := execCommand("gh", args...)
	cmdCreate.Dir = workDir
	var outCreate bytes.Buffer
	var errCreate bytes.Buffer
	cmdCreate.Stdout = &outCreate
	cmdCreate.Stderr = &errCreate
	err = cmdCreate.Run()

	if err != nil {
		return "", fmt.Errorf("gh pr create failed: %s: %w", errCreate.String(), err)
	}

	prURL := strings.TrimSpace(outCreate.String())
	return prURL, nil
}
