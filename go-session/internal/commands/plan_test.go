package commands_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	commands "github.com/daniel-talonone/gemini-commands/internal/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const samplePlan = `- id: slice-one
  description: First slice
  status: todo
  tasks:
    - id: task-one
      task: do the thing
      status: todo
    - id: task-two
      task: do another thing
      status: todo
- id: slice-two
  description: Second slice
  status: todo
  tasks:
    - id: task-three
      task: yet another thing
      status: todo
`

func writePlan(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte(content), 0644))
	return dir
}

func readPlan(t *testing.T, dir string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, "plan.yml"))
	require.NoError(t, err)
	return string(content)
}

func TestUpdateTask_SetsDone(t *testing.T) {
	dir := writePlan(t, samplePlan)
	require.NoError(t, commands.UpdateTask(dir, "task-one", "done"))
	plan := readPlan(t, dir)
	assert.Contains(t, plan, "status: done")
}

func TestUpdateTask_SetsInProgress(t *testing.T) {
	dir := writePlan(t, samplePlan)
	require.NoError(t, commands.UpdateTask(dir, "task-two", "in-progress"))
	assert.Contains(t, readPlan(t, dir), "status: in-progress")
}

func TestUpdateTask_PreservesOtherContent(t *testing.T) {
	dir := writePlan(t, samplePlan)
	require.NoError(t, commands.UpdateTask(dir, "task-one", "done"))
	plan := readPlan(t, dir)
	assert.Contains(t, plan, "id: task-two")
	assert.Contains(t, plan, "id: slice-two")
	assert.Contains(t, plan, "do the thing")
	assert.Contains(t, plan, "yet another thing")
	assert.Equal(t, 1, strings.Count(plan, "status: done"), "only task-one should be done")
}

func TestUpdateTask_TaskNotFound(t *testing.T) {
	dir := writePlan(t, samplePlan)
	err := commands.UpdateTask(dir, "nonexistent", "done")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestUpdateTask_InvalidStatus(t *testing.T) {
	dir := writePlan(t, samplePlan)
	err := commands.UpdateTask(dir, "task-one", "invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestUpdateTask_MissingPlanYml(t *testing.T) {
	dir := t.TempDir()
	err := commands.UpdateTask(dir, "task-one", "done")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan.yml")
}

func TestUpdateSlice_SetsDone(t *testing.T) {
	dir := writePlan(t, samplePlan)
	require.NoError(t, commands.UpdateSlice(dir, "slice-one", "done"))
	plan := readPlan(t, dir)
	assert.Contains(t, plan, "status: done")
	assert.Contains(t, plan, "id: slice-one")
}

func TestUpdateSlice_PreservesOtherContent(t *testing.T) {
	dir := writePlan(t, samplePlan)
	require.NoError(t, commands.UpdateSlice(dir, "slice-one", "done"))
	plan := readPlan(t, dir)
	assert.Contains(t, plan, "id: slice-two")
	assert.Equal(t, 1, strings.Count(plan, "status: done"))
}

func TestUpdateSlice_SliceNotFound(t *testing.T) {
	dir := writePlan(t, samplePlan)
	err := commands.UpdateSlice(dir, "nonexistent-slice", "done")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-slice")
}
