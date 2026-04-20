package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
      status: in-progress
- id: slice-two
  description: Second slice
  status: done
  tasks:
    - id: task-three
      task: yet another thing
      status: done
`

func writePlanFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte(samplePlan), 0644))
	return dir
}

func readPlanFile(t *testing.T, dir string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, "plan.yml"))
	require.NoError(t, err)
	return string(content)
}

func TestResetPlan_ResetsAllStatuses(t *testing.T) {
	dir := writePlanFile(t)
	require.NoError(t, ResetPlan(dir))
	content := readPlanFile(t, dir)

	// Count "status: todo" occurrences - should be 3 slices + 3 tasks = at least 3 visible
	todoCount := strings.Count(content, "status: todo")
	assert.GreaterOrEqual(t, todoCount, 3, "should have at least 3 'todo' statuses")

	// Verify no "done" statuses remain
	assert.NotContains(t, content, "status: done", "should not have any 'done' statuses")

	// Verify no "in-progress" statuses remain
	assert.NotContains(t, content, "status: in-progress", "should not have any 'in-progress' statuses")
}

func TestResetPlan_PreservesSliceIDs(t *testing.T) {
	dir := writePlanFile(t)
	require.NoError(t, ResetPlan(dir))
	content := readPlanFile(t, dir)

	assert.Contains(t, content, "id: slice-one")
	assert.Contains(t, content, "id: slice-two")
}

func TestResetPlan_PreservesTaskIDs(t *testing.T) {
	dir := writePlanFile(t)
	require.NoError(t, ResetPlan(dir))
	content := readPlanFile(t, dir)

	assert.Contains(t, content, "id: task-one")
	assert.Contains(t, content, "id: task-two")
	assert.Contains(t, content, "id: task-three")
}

func TestResetPlan_PreservesDescriptions(t *testing.T) {
	dir := writePlanFile(t)
	require.NoError(t, ResetPlan(dir))
	content := readPlanFile(t, dir)

	assert.Contains(t, content, "description: First slice")
	assert.Contains(t, content, "description: Second slice")
	assert.Contains(t, content, "task: do the thing")
	assert.Contains(t, content, "task: do another thing")
	assert.Contains(t, content, "task: yet another thing")
}

func TestResetPlan_MissingPlanYml(t *testing.T) {
	dir := t.TempDir()
	err := ResetPlan(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan.yml")
}

func TestResetPlan_EmptyPlanHandling(t *testing.T) {
	// Test that ResetPlan handles empty plan gracefully
	// Note: LoadPlan will succeed with an empty slice, but ValidatePlan should catch it
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte("[]"), 0644))

	// This should fail validation due to "plan must contain at least one slice"
	err := ResetPlan(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan must contain at least one slice")
}
