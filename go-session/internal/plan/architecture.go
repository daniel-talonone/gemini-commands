package plan

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadArchitecture reads architecture.md from featureDir. Returns an empty string
// without error if the file does not exist — architecture is optional.
func LoadArchitecture(featureDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(featureDir, "architecture.md"))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading architecture.md: %w", err)
	}
	return string(data), nil
}

// WriteArchitecture saves the architecture content to architecture.md in an atomic way.
func WriteArchitecture(featureDir string, content string) error {
	p := filepath.Join(featureDir, "architecture.md")
	tempFile, err := os.CreateTemp(featureDir, "architecture.md.*")
	if err != nil {
		return fmt.Errorf("creating temp file for architecture: %w", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	if _, err := tempFile.WriteString(content); err != nil {
		return fmt.Errorf("writing to temp file for architecture: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("closing temp file for architecture: %w", err)
	}

	if err := os.Rename(tempFile.Name(), p); err != nil {
		return fmt.Errorf("renaming temp file for architecture: %w", err)
	}

	return nil
}
