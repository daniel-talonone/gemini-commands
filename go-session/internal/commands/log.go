package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AppendLog appends a timestamped Markdown entry to log.md in featureDir.
// Creates log.md if it does not exist.
// Returns an error if featureDir does not exist.
func AppendLog(featureDir, message string) error {
	if _, err := os.Stat(featureDir); os.IsNotExist(err) {
		return fmt.Errorf("feature directory does not exist: %s", featureDir)
	}

	logPath := filepath.Join(featureDir, "log.md")
	timestamp := time.Now().UTC().Format(time.RFC3339)
	entry := fmt.Sprintf("## [%s]\n\n%s\n", timestamp, message)

	existing, err := os.ReadFile(logPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading log.md: %w", err)
	}

	var content string
	if len(existing) > 0 {
		content = string(existing) + "\n" + entry
	} else {
		content = entry
	}

	return os.WriteFile(logPath, []byte(content), 0644)
}
