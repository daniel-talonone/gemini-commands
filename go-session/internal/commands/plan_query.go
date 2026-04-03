package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SliceSummary is the lightweight view of a slice (no task bodies).
type SliceSummary struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Status      string   `yaml:"status"`
	DependsOn   []string `yaml:"depends_on"`
}

// TaskSummary is the lightweight view of a task (ID + status only).
type TaskSummary struct {
	ID     string `yaml:"id"`
	Status string `yaml:"status"`
}

// Task is the full view of a single task.
type Task struct {
	ID          string `yaml:"id"`
	Task        string `yaml:"task"`
	Description string `yaml:"description"`
	Status      string `yaml:"status"`
}

// slice is the internal full representation used for parsing.
type slice struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Status      string   `yaml:"status"`
	DependsOn   []string `yaml:"depends_on"`
	Tasks       []Task   `yaml:"tasks"`
}

func loadPlan(featureDir string) ([]slice, error) {
	planPath := filepath.Join(featureDir, "plan.yml")
	data, err := os.ReadFile(planPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plan.yml not found in %s", featureDir)
		}
		return nil, fmt.Errorf("reading plan.yml: %w", err)
	}
	var slices []slice
	if err := yaml.Unmarshal(data, &slices); err != nil {
		return nil, fmt.Errorf("parsing plan.yml: %w", err)
	}
	return slices, nil
}

// ListSlices returns a summary (id + status) for every slice in plan.yml.
func ListSlices(featureDir string) ([]SliceSummary, error) {
	slices, err := loadPlan(featureDir)
	if err != nil {
		return nil, err
	}
	out := make([]SliceSummary, len(slices))
	for i, s := range slices {
		out[i] = SliceSummary{
			ID:          s.ID,
			Description: s.Description,
			Status:      s.Status,
			DependsOn:   s.DependsOn,
		}
	}
	return out, nil
}

// ListTasks returns a lightweight summary (id + status) for every task in the given slice.
func ListTasks(featureDir, sliceID string) ([]TaskSummary, error) {
	slices, err := loadPlan(featureDir)
	if err != nil {
		return nil, err
	}
	for _, s := range slices {
		if s.ID == sliceID {
			out := make([]TaskSummary, len(s.Tasks))
			for i, t := range s.Tasks {
				out[i] = TaskSummary{ID: t.ID, Status: t.Status}
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("slice %q not found in plan.yml", sliceID)
}

// GetSlice returns the full details of a slice, without task bodies.
func GetSlice(featureDir, sliceID string) (*SliceSummary, error) {
	slices, err := loadPlan(featureDir)
	if err != nil {
		return nil, err
	}
	for _, s := range slices {
		if s.ID == sliceID {
			return &SliceSummary{
				ID:          s.ID,
				Description: s.Description,
				Status:      s.Status,
				DependsOn:   s.DependsOn,
			}, nil
		}
	}
	return nil, fmt.Errorf("slice %q not found in plan.yml", sliceID)
}

// GetTask returns the full content of a single task within a slice.
func GetTask(featureDir, sliceID, taskID string) (*Task, error) {
	slices, err := loadPlan(featureDir)
	if err != nil {
		return nil, err
	}
	for _, s := range slices {
		if s.ID != sliceID {
			continue
		}
		for _, t := range s.Tasks {
			if t.ID == taskID {
				return &Task{
					ID:          t.ID,
					Task:        t.Task,
					Description: t.Description,
					Status:      t.Status,
				}, nil
			}
		}
		return nil, fmt.Errorf("task %q not found in slice %q", taskID, sliceID)
	}
	return nil, fmt.Errorf("slice %q not found in plan.yml", sliceID)
}
