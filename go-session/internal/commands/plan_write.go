package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var kebabRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type planSliceRaw struct {
	ID          string        `yaml:"id"`
	Description string        `yaml:"description"`
	Status      string        `yaml:"status"`
	Tasks       []planTaskRaw `yaml:"tasks"`
}

type planTaskRaw struct {
	ID     string `yaml:"id"`
	Task   string `yaml:"task"`
	Status string `yaml:"status"`
}

// ValidatePlan parses data and enforces the full plan schema.
// Returns a precise, location-qualified error message on any violation.
func ValidatePlan(data []byte) error {
	var slices []planSliceRaw
	if err := yaml.Unmarshal(data, &slices); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}
	if len(slices) == 0 {
		return fmt.Errorf("plan must contain at least one slice")
	}

	seenSliceIDs := map[string]bool{}
	seenTaskIDs := map[string]string{} // taskID → sliceID

	for _, s := range slices {
		loc := fmt.Sprintf("slice %q", s.ID)
		if s.ID == "" {
			return fmt.Errorf("a slice is missing field \"id\"")
		}
		if !kebabRe.MatchString(s.ID) {
			return fmt.Errorf("%s: id is not kebab-case", loc)
		}
		if seenSliceIDs[s.ID] {
			return fmt.Errorf("duplicate slice id %q", s.ID)
		}
		seenSliceIDs[s.ID] = true
		if strings.TrimSpace(s.Description) == "" {
			return fmt.Errorf("%s: missing field \"description\"", loc)
		}
		if !validStatuses[s.Status] {
			return fmt.Errorf("%s: invalid status %q — must be todo, in-progress, or done", loc, s.Status)
		}
		if len(s.Tasks) == 0 {
			return fmt.Errorf("%s: tasks must be a non-empty list", loc)
		}
		for _, t := range s.Tasks {
			tloc := fmt.Sprintf("slice %q: task %q", s.ID, t.ID)
			if t.ID == "" {
				return fmt.Errorf("%s: a task is missing field \"id\"", loc)
			}
			if !kebabRe.MatchString(t.ID) {
				return fmt.Errorf("%s: id is not kebab-case", tloc)
			}
			if prev, exists := seenTaskIDs[t.ID]; exists {
				return fmt.Errorf("duplicate task id %q (found in slice %q and %q)", t.ID, prev, s.ID)
			}
			seenTaskIDs[t.ID] = s.ID
			if strings.TrimSpace(t.Task) == "" {
				return fmt.Errorf("%s: missing field \"task\"", tloc)
			}
			if !validStatuses[t.Status] {
				return fmt.Errorf("%s: invalid status %q — must be todo, in-progress, or done", tloc, t.Status)
			}
		}
	}
	return nil
}

// WritePlan validates data against the plan schema and writes it atomically
// to plan.yml in featureDir. The original bytes are preserved — no reformatting.
func WritePlan(featureDir string, data []byte) error {
	if err := ValidatePlan(data); err != nil {
		return err
	}
	planPath := filepath.Join(featureDir, "plan.yml")
	tmpPath := planPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("writing plan.yml.tmp: %w", err)
	}
	return os.Rename(tmpPath, planPath)
}
