package plan

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadQuestions reads questions.yml from featureDir. Returns an empty string
// without error if the file does not exist — questions are optional.
func LoadQuestions(featureDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(featureDir, "questions.yml"))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading questions.yml: %w", err)
	}
	return string(data), nil
}
