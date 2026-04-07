package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"gopkg.in/yaml.v3"
)

// SplitTaskEntry is one replacement task produced by the enricher's SPLIT output.
type SplitTaskEntry struct {
	Suffix string `yaml:"suffix"`
	Task   string `yaml:"task"`
}

// SplitTask replaces a single todo task with N atomic subtasks.
// Generated IDs are "{taskID}-{suffix}" for each entry.
//
// Validation order:
//  1. At least 2 replacements
//  2. Each suffix is kebab-case
//  3. Each task body passes the injection guard (no id:/status: lines)
//  4. Load plan.yml
//  5. All generated IDs are unique across the entire plan (and among themselves)
//  6. Slice + task exist; task status is "todo"
//  7. Splice replacement nodes in place of the original
//  8. Atomic write
func SplitTask(featureDir, sliceID, taskID string, replacements []SplitTaskEntry) error {
	// 1. Minimum replacement count
	if len(replacements) < 2 {
		return fmt.Errorf("must provide at least 2 replacement tasks, got %d", len(replacements))
	}

	// 2 & 3. Validate suffixes and task bodies
	for i, r := range replacements {
		if !plan.KebabRe.MatchString(r.Suffix) {
			return fmt.Errorf("replacement %d: suffix %q is not kebab-case", i, r.Suffix)
		}
		for _, line := range strings.Split(r.Task, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "id:") {
				return fmt.Errorf("replacement %d (suffix %q): task body must not contain \"id:\"", i, r.Suffix)
			}
			if strings.HasPrefix(trimmed, "status:") {
				return fmt.Errorf("replacement %d (suffix %q): task body must not contain \"status:\"", i, r.Suffix)
			}
		}
	}

	// 4. Load plan
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

	// 5. Collect all existing task IDs (excluding the task being replaced)
	existingTaskIDs := map[string]string{} // id → sliceID
	for _, sNode := range sliceSeq.Content {
		if sNode.Kind != yaml.MappingNode {
			continue
		}
		sID := plan.MappingScalar(sNode, "id")
		tasksNode := plan.MappingValue(sNode, "tasks")
		if tasksNode == nil || tasksNode.Kind != yaml.SequenceNode {
			continue
		}
		for _, tNode := range tasksNode.Content {
			if tNode.Kind != yaml.MappingNode {
				continue
			}
			tID := plan.MappingScalar(tNode, "id")
			if tID != "" && (sID != sliceID || tID != taskID) {
				existingTaskIDs[tID] = sID
			}
		}
	}

	// Check generated IDs for collisions (against existing tasks and among themselves)
	seenGenerated := map[string]bool{}
	for _, r := range replacements {
		genID := taskID + "-" + r.Suffix
		if inSlice, exists := existingTaskIDs[genID]; exists {
			return fmt.Errorf("generated id %q collides with existing task id in slice %q", genID, inSlice)
		}
		if seenGenerated[genID] {
			return fmt.Errorf("generated id %q collides with another replacement (duplicate suffix %q)", genID, r.Suffix)
		}
		seenGenerated[genID] = true
	}

	// 6. Find slice → find task → verify status
	sliceFound := false
	taskFound := false

	for _, sNode := range sliceSeq.Content {
		if sNode.Kind != yaml.MappingNode || plan.MappingScalar(sNode, "id") != sliceID {
			continue
		}
		sliceFound = true

		tasksNode := plan.MappingValue(sNode, "tasks")
		if tasksNode == nil || tasksNode.Kind != yaml.SequenceNode {
			return fmt.Errorf("slice %q: tasks field is missing or invalid", sliceID)
		}

		targetIdx := -1
		for i, tNode := range tasksNode.Content {
			if tNode.Kind != yaml.MappingNode || plan.MappingScalar(tNode, "id") != taskID {
				continue
			}
			taskFound = true
			status := plan.MappingScalar(tNode, "status")
			if status != "todo" {
				return fmt.Errorf("task %q has status %q — split skipped (only todo tasks may be split)", taskID, status)
			}
			targetIdx = i
			break
		}

		if !taskFound {
			break
		}

		// 7. Build replacement nodes and splice
		newNodes := make([]*yaml.Node, 0, len(replacements))
		for _, r := range replacements {
			newNodes = append(newNodes, buildSplitTaskNode(taskID+"-"+r.Suffix, r.Task))
		}

		before := make([]*yaml.Node, targetIdx)
		copy(before, tasksNode.Content[:targetIdx])
		after := tasksNode.Content[targetIdx+1:]
		tasksNode.Content = append(append(before, newNodes...), after...)
		break
	}

	if !sliceFound {
		return fmt.Errorf("slice %q not found in plan.yml", sliceID)
	}
	if !taskFound {
		return fmt.Errorf("task %q not found in slice %q", taskID, sliceID)
	}

	// 8. Atomic write
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

// buildSplitTaskNode creates a yaml.MappingNode for a new task with status todo.
func buildSplitTaskNode(id, task string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	node.Content = []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "id", Tag: "!!str"},
		{Kind: yaml.ScalarNode, Value: id, Tag: "!!str"},
		{Kind: yaml.ScalarNode, Value: "task", Tag: "!!str"},
		{Kind: yaml.ScalarNode, Value: task, Tag: "!!str"},
		{Kind: yaml.ScalarNode, Value: "status", Tag: "!!str"},
		{Kind: yaml.ScalarNode, Value: "todo", Tag: "!!str"},
	}
	return node
}
