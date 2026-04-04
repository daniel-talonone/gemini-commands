package commands

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type statusFile struct {
	Mode         string `yaml:"mode"`
	Repo         string `yaml:"repo"`
	Branch       string `yaml:"branch"`
	PID          int    `yaml:"pid"`
	PipelineStep string `yaml:"pipeline_step"`
	StartedAt    string `yaml:"started_at"`
	UpdatedAt    string `yaml:"updated_at"`
}

// updateStatusPipelineStep updates pipeline_step and updated_at in status.yaml.
// Best-effort: silently does nothing if status.yaml does not exist or cannot be parsed.
func updateStatusPipelineStep(featureDir, step string) {
	statusPath := filepath.Join(featureDir, "status.yaml")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return // missing or unreadable — nothing to update
	}
	var s statusFile
	if err := yaml.Unmarshal(data, &s); err != nil {
		return
	}
	s.PipelineStep = step
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	out, err := yaml.Marshal(&s)
	if err != nil {
		return
	}
	_ = os.WriteFile(statusPath, out, 0644)
}
