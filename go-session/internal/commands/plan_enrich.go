package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnrichTask updates only the task: body of a single todo task in plan.yml
// using the yaml.Node API. All other fields are preserved exactly.
//
// Validation order:
//  1. body non-empty
//  2. body has no lines starting with "id:" or "status:" (injection guard)
//  3. slice + task exist
//  4. task status is "todo" (done/in-progress are protected)
func EnrichTask(featureDir, sliceID, taskID, body string) error {
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("task body must not be empty")
	}
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "id:") {
			return fmt.Errorf("task body must not contain \"id:\" — pass only the task description text, not YAML fields")
		}
		if strings.HasPrefix(trimmed, "status:") {
			return fmt.Errorf("task body must not contain \"status:\" — pass only the task description text, not YAML fields")
		}
	}

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
		return fmt.Errorf("plan.yml: expected top-level sequence")
	}

	sliceFound := false
	taskFound := false

	for _, sliceNode := range sliceSeq.Content {
		if sliceNode.Kind != yaml.MappingNode {
			continue
		}
		if mappingScalar(sliceNode, "id") != sliceID {
			continue
		}
		sliceFound = true

		tasksNode := mappingValue(sliceNode, "tasks")
		if tasksNode == nil || tasksNode.Kind != yaml.SequenceNode {
			return fmt.Errorf("slice %q: tasks field is missing or invalid", sliceID)
		}

		for _, taskNode := range tasksNode.Content {
			if taskNode.Kind != yaml.MappingNode {
				continue
			}
			if mappingScalar(taskNode, "id") != taskID {
				continue
			}
			taskFound = true

			status := mappingScalar(taskNode, "status")
			if status != "todo" {
				return fmt.Errorf("task %q has status %q — enrichment skipped (only todo tasks may be enriched)", taskID, status)
			}

			if err := setMappingValue(taskNode, "task", body); err != nil {
				return fmt.Errorf("updating task %q in slice %q: %w", taskID, sliceID, err)
			}
			break
		}
		break
	}

	if !sliceFound {
		return fmt.Errorf("slice %q not found in plan.yml", sliceID)
	}
	if !taskFound {
		return fmt.Errorf("task %q not found in slice %q", taskID, sliceID)
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
