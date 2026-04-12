package pr

import (
	"fmt"
	"os"
	"path/filepath"
)

const filename = "pr.md"

// Create initializes pr.md with a placeholder in the specified feature directory.
// Does nothing if the file already exists.
func Create(featureDir string) error {
	filePath := filepath.Join(featureDir, filename)
	if _, err := os.Stat(filePath); err == nil {
		return nil // File exists, nothing to do
	}
	return os.WriteFile(filePath, []byte("# Pull Request\n"), 0644)
}

// Write saves the given content to pr.md in the specified feature directory.
// It uses a temporary file and rename to ensure atomic writes.
func Write(featureDir, content string) error {
	filePath := filepath.Join(featureDir, filename)
	tmpFilePath := filePath + ".tmp"

	if err := os.WriteFile(tmpFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temporary PR file: %w", err)
	}

	if err := os.Rename(tmpFilePath, filePath); err != nil {
		return fmt.Errorf("failed to rename temporary PR file: %w", err)
	}
	return nil
}

// Read reads the content of pr.md from the specified feature directory.
func Read(featureDir string) (string, error) {
	filePath := filepath.Join(featureDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Return empty string if file does not exist
		}
		return "", fmt.Errorf("failed to read PR file: %w", err)
	}
	return string(content), nil
}
