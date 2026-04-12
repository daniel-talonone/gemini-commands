package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

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
	cmd := exec.Command("gh", "pr", "view", "--json", "reviewThreads")
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
