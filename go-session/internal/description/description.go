package description

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
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

// RenderMarkdown converts a markdown string to safe HTML using goldmark.
// Returns empty template.HTML if input is empty or rendering fails.
func RenderMarkdown(markdown string) template.HTML {
	if markdown == "" {
		return template.HTML("")
	}
	var buf bytes.Buffer
	if err := goldmark.New().Convert([]byte(markdown), &buf); err != nil {
		return template.HTML("")
	}
	return template.HTML(buf.String())
}
