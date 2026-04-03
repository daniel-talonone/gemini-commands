package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const queryPlan = `- id: slice-a
  description: First slice
  status: done
  depends_on: []
  tasks:
    - id: task-one
      task: do the thing
      status: done
    - id: task-two
      task: do another thing
      status: in-progress
- id: slice-b
  description: Second slice
  status: todo
  depends_on:
    - slice-a
  tasks:
    - id: task-three
      task: yet another thing
      status: todo
`

func writeQueryPlan(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte(queryPlan), 0644))
	return dir
}

func TestListSlices(t *testing.T) {
	dir := writeQueryPlan(t)
	slices, err := commands.ListSlices(dir)
	require.NoError(t, err)
	require.Len(t, slices, 2)
	assert.Equal(t, "slice-a", slices[0].ID)
	assert.Equal(t, "done", slices[0].Status)
	assert.Equal(t, "slice-b", slices[1].ID)
	assert.Equal(t, "todo", slices[1].Status)
}

func TestListSlices_MissingPlan(t *testing.T) {
	_, err := commands.ListSlices(t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan.yml")
}

func TestListTasks(t *testing.T) {
	dir := writeQueryPlan(t)
	tasks, err := commands.ListTasks(dir, "slice-a")
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	assert.Equal(t, "task-one", tasks[0].ID)
	assert.Equal(t, "done", tasks[0].Status)
	assert.Equal(t, "task-two", tasks[1].ID)
	assert.Equal(t, "in-progress", tasks[1].Status)
}

func TestListTasks_SliceNotFound(t *testing.T) {
	dir := writeQueryPlan(t)
	_, err := commands.ListTasks(dir, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestGetSlice(t *testing.T) {
	dir := writeQueryPlan(t)
	s, err := commands.GetSlice(dir, "slice-b")
	require.NoError(t, err)
	assert.Equal(t, "slice-b", s.ID)
	assert.Equal(t, "Second slice", s.Description)
	assert.Equal(t, []string{"slice-a"}, s.DependsOn)
}

func TestGetSlice_NotFound(t *testing.T) {
	dir := writeQueryPlan(t)
	_, err := commands.GetSlice(dir, "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestGetTask(t *testing.T) {
	dir := writeQueryPlan(t)
	task, err := commands.GetTask(dir, "slice-a", "task-two")
	require.NoError(t, err)
	assert.Equal(t, "task-two", task.ID)
	assert.Equal(t, "in-progress", task.Status)
	assert.Contains(t, task.Task, "do another thing")
}

func TestGetTask_TaskNotFound(t *testing.T) {
	dir := writeQueryPlan(t)
	_, err := commands.GetTask(dir, "slice-a", "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestGetTask_SliceNotFound(t *testing.T) {
	dir := writeQueryPlan(t)
	_, err := commands.GetTask(dir, "nonexistent", "task-one")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}
