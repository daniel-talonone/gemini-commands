package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var validStatuses = map[string]bool{
	"todo": true, "in-progress": true, "done": true,
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
	planPath := filepath.Join(featureDir, "plan.yml")
	data, err := os.ReadFile(planPath)
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
			tasksSeq := mappingValue(sliceNode, "tasks")
			if tasksSeq == nil || tasksSeq.Kind != yaml.SequenceNode {
				continue
			}
			for _, taskNode := range tasksSeq.Content {
				if taskNode.Kind != yaml.MappingNode {
					continue
				}
				if mappingScalar(taskNode, "id") == id {
					if err := setMappingValue(taskNode, "status", status); err != nil {
						return err
					}
					found = true
					break
				}
			}
		} else {
			if mappingScalar(sliceNode, "id") == id {
				if err := setMappingValue(sliceNode, "status", status); err != nil {
					return err
				}
				found = true
			}
		}
		if found {
			break
		}
	}

	if !found {
		kind := "task"
		if !isTask {
			kind = "slice"
		}
		return fmt.Errorf("%s %q not found in plan.yml", kind, id)
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("marshaling plan.yml: %w", err)
	}

	tmpPath := planPath + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0644); err != nil {
		return fmt.Errorf("writing plan.yml.tmp: %w", err)
	}
	return os.Rename(tmpPath, planPath)
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func mappingScalar(node *yaml.Node, key string) string {
	v := mappingValue(node, key)
	if v == nil {
		return ""
	}
	return v.Value
}

func setMappingValue(node *yaml.Node, key, value string) error {
	v := mappingValue(node, key)
	if v == nil {
		return fmt.Errorf("key %q not found in mapping", key)
	}
	v.Value = value
	return nil
}
