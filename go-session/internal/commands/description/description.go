package description

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadDescription reads description.md from featureDir.
func LoadDescription(featureDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(featureDir, "description.md"))
	if err != nil {
		return "", fmt.Errorf("reading description.md: %w", err)
	}
	return string(data), nil
}

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
