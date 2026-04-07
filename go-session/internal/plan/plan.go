package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/daniel-talonone/gemini-commands/internal/status"
	"gopkg.in/yaml.v3"
)

// KebabRe matches a valid kebab-case identifier (e.g. "my-slice-id").
var KebabRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var validStatuses = map[string]bool{
	"todo": true, "in-progress": true, "done": true,
}

// Plan is a collection of Slices, representing the entire plan.yml file.
type Plan []Slice

// Slice represents a single slice in the plan.
type Slice struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Status      string   `yaml:"status"`
	DependsOn   []string `yaml:"depends_on,omitempty"`
	Tasks       []Task   `yaml:"tasks"`
}

// Task represents a single task within a slice.
type Task struct {
	ID     string `yaml:"id"`
	Task   string `yaml:"task"`
	Status string `yaml:"status"`
}

// LoadPlan reads and parses plan.yml from featureDir into a Plan.
func LoadPlan(featureDir string) (Plan, error) {
	data, err := os.ReadFile(planPath(featureDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plan.yml not found in %s", featureDir)
		}
		return nil, fmt.Errorf("reading plan.yml: %w", err)
	}
	var p Plan
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing plan.yml: %w", err)
	}
	return p, nil
}

// ValidatePlan parses data and enforces the full plan schema.
// Returns a precise, location-qualified error message on any violation.
func ValidatePlan(data []byte) error {
	var slices []Slice
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
			return fmt.Errorf("a slice is missing field 'id'") // Corrected
		}
		if !KebabRe.MatchString(s.ID) {
			return fmt.Errorf("%s: id is not kebab-case", loc)
		}
		if seenSliceIDs[s.ID] {
			return fmt.Errorf("duplicate slice id %q", s.ID)
		}
		seenSliceIDs[s.ID] = true
		if strings.TrimSpace(s.Description) == "" {
			return fmt.Errorf("%s: missing field 'description'", loc) // Corrected
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
				return fmt.Errorf("%s: a task is missing field 'id'", loc) // Corrected
			}
			if !KebabRe.MatchString(t.ID) {
				return fmt.Errorf("%s: id is not kebab-case", tloc)
			}
			if prev, exists := seenTaskIDs[t.ID]; exists {
				return fmt.Errorf("duplicate task id %q (found in slice %q and %q)", t.ID, prev, s.ID)
			}
			seenTaskIDs[t.ID] = s.ID
			if strings.TrimSpace(t.Task) == "" {
				return fmt.Errorf("%s: missing field 'task'", tloc) // Corrected
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
	var err error
	if err = ValidatePlan(data); err != nil {
		return err
	}
	tmpPath := planPath(featureDir) + ".tmp"
	if err = os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("writing plan.yml.tmp: %w", err)
	}
	if err = os.Rename(tmpPath, planPath(featureDir)); err != nil {
		return err
	}
	if err = status.Write(featureDir, "plan-done", "", ""); err != nil {
		return fmt.Errorf("updating status pipeline step: %w", err)
	}
	return nil
}

// UpdateTask updates the status of a task (nested inside a slice) in plan.yml.
// Preserves all other YAML content exactly via the yaml.Node API.
func UpdateTask(featureDir, taskID, status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("invalid status %q: must be one of: todo, in-progress, done", status)
	}
	return updatePlanStatus(featureDir, taskID, status, true)
}

// UpdateSlice updates the status of a top-level slice in plan.yml.
// Preserves all other YAML content exactly via the yaml.Node API.
func UpdateSlice(featureDir, sliceID, status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("invalid status %q: must be one of: todo, in-progress, done", status)
	}
	return updatePlanStatus(featureDir, sliceID, status, false)
}

func updatePlanStatus(featureDir, id, status string, isTask bool) error {
	data, err := os.ReadFile(planPath(featureDir))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plan.yml not found in %s", featureDir)
		}
		return fmt.Errorf("reading plan.yml: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parsing plan.yml: %w", err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return fmt.Errorf("plan.yml: expected document node with content")
	}

	sliceSeq := doc.Content[0]
	if sliceSeq.Kind != yaml.SequenceNode {
		return fmt.Errorf("plan.yml: expected top-level sequence, got kind %d", sliceSeq.Kind)
	}

	found := false
	for _, sliceNode := range sliceSeq.Content {
		if sliceNode.Kind != yaml.MappingNode {
			continue
		}
		if isTask {
			// Search for task
			tasksNode := MappingValue(sliceNode, "tasks")
			if tasksNode == nil || tasksNode.Kind != yaml.SequenceNode {
				continue
			}
			for _, taskNode := range tasksNode.Content {
				if taskNode.Kind != yaml.MappingNode {
					continue
				}
				if MappingScalar(taskNode, "id") == id {
					SetMappingValue(taskNode, "status", status)
					found = true
					break
				}
			}
		} else {
			// Search for slice
			if MappingScalar(sliceNode, "id") == id {
				SetMappingValue(sliceNode, "status", status)
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("%s with id %q not found in plan.yml",
			func() string {
				if isTask {
					return "task"
				}
				return "slice"
			}(), id)
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("marshaling plan.yml: %w", err)
	}

	tmpPath := planPath(featureDir) + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0644); err != nil {
		return fmt.Errorf("writing plan.yml.tmp: %w", err)
	}
	return os.Rename(tmpPath, planPath(featureDir))
}

// planPath returns the full path to plan.yml for a given feature directory.
func planPath(featureDir string) string {
	return filepath.Join(featureDir, "plan.yml")
}

// Helper functions for yaml.Node manipulation.

// MappingScalar returns the scalar value of a given key from a YAML mapping node.
func MappingScalar(node *yaml.Node, key string) string {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1].Value
		}
	}
	return ""
}

// MappingValue returns the yaml.Node of a given key from a YAML mapping node.
func MappingValue(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// SetMappingValue sets the value of a given key in a YAML mapping node.
// If the key does not exist, it adds it.
func SetMappingValue(node *yaml.Node, key, value string) {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			node.Content[i+1].SetString(value)
			return
		}
	}
	// Key not found, add it
	node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, &yaml.Node{Kind: yaml.ScalarNode, Value: value})
}
