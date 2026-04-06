package status

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Status represents the structure of the status.yaml file.
type Status struct {
	Mode         string `yaml:"mode"`
	Repo         string `yaml:"repo"`
	Branch       string `yaml:"branch"`
	WorkDir      string `yaml:"work_dir"`
	PID          int    `yaml:"pid"`
	PipelineStep string `yaml:"pipeline_step"`
	StartedAt    string `yaml:"started_at"`
	UpdatedAt    string `yaml:"updated_at"`
}

// Write reads, updates, and atomically writes the status.yaml file.
// If repo or branch are empty, their existing values are preserved.
func Write(featureDir, step, repo, branch string) error {
	statusPath := filepath.Join(featureDir, "status.yaml")
	var s Status

	// Read existing status.yaml if it exists
	data, err := os.ReadFile(statusPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading status.yaml: %w", err)
		}
		// File does not exist, initialize a new Status struct
		s = Status{
			Mode:      "auto", // Default mode
			StartedAt: time.Now().Format(time.RFC3339),
		}
	} else {
		// File exists, unmarshal existing data
		if err := yaml.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("unmarshaling status.yaml: %w", err)
		}
	}

	// Update fields
	s.PipelineStep = step
	s.UpdatedAt = time.Now().Format(time.RFC3339)

	// Only update repo/branch if provided, otherwise preserve existing values
	if repo != "" {
		s.Repo = repo
	}
	if branch != "" {
		s.Branch = branch
	}

	// Ensure StartedAt is set if it wasn't already (e.g., for existing files without it)
	if s.StartedAt == "" {
		s.StartedAt = s.UpdatedAt // Use UpdatedAt as fallback for StartedAt if missing
	}

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(&s)
	if err != nil {
		return fmt.Errorf("marshaling status.yaml: %w", err)
	}

	// Atomically write to file
	tmpPath := statusPath + ".tmp"
	if err := os.WriteFile(tmpPath, updatedData, 0644); err != nil {
		return fmt.Errorf("writing status.yaml.tmp: %w", err)
	}
	if err := os.Rename(tmpPath, statusPath); err != nil {
		return fmt.Errorf("renaming status.yaml.tmp to status.yaml: %w", err)
	}

	return nil
}