package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadContext reads all .md, .yml, .yaml files from featureDir (excluding _* files),
// sorts them alphabetically, and returns them formatted as XML blocks for LLM consumption.
//
// Output format:
//
//	<file name="description.md">
//	...content...
//	</file>
//
//	<file name="plan.yml">
//	...content...
//	</file>
func LoadContext(featureDir string) (string, error) {
	if _, err := os.Stat(featureDir); os.IsNotExist(err) {
		return "", fmt.Errorf("feature directory does not exist: %s", featureDir)
	}

	entries, err := os.ReadDir(featureDir)
	if err != nil {
		return "", fmt.Errorf("reading feature directory: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "_") {
			continue
		}
		ext := filepath.Ext(name)
		if ext != ".md" && ext != ".yml" && ext != ".yaml" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	var blocks []string
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(featureDir, name))
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", name, err)
		}
		blocks = append(blocks, fmt.Sprintf("<file name=%q>\n%s\n</file>", name, content))
	}

	return strings.Join(blocks, "\n\n"), nil
}
