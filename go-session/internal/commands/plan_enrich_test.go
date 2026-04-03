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

const enrichTestPlan = `- id: slice-a
  description: First slice
  status: todo
  tasks:
    - id: task-todo
      task: old body
      status: todo
    - id: task-in-progress
      task: wip body
      status: in-progress
    - id: task-done
      task: done body
      status: done
- id: slice-b
  description: Second slice
  status: todo
  tasks:
    - id: task-other
      task: untouched body
      status: todo
`

func writeEnrichPlan(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte(enrichTestPlan), 0644))
}

func readEnrichPlan(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "plan.yml"))
	require.NoError(t, err)
	return string(b)
}

func TestEnrichTask_UpdatesBody(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)

	newBody := "FILE: src/foo.go — new enriched description"
	require.NoError(t, commands.EnrichTask(dir, "slice-a", "task-todo", newBody))

	plan := readEnrichPlan(t, dir)
	assert.Contains(t, plan, newBody)
	assert.Contains(t, plan, "task-todo")   // id preserved
	assert.Contains(t, plan, "status: todo") // status preserved
}

func TestEnrichTask_OtherTasksUntouched(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)

	require.NoError(t, commands.EnrichTask(dir, "slice-a", "task-todo", "new body"))

	plan := readEnrichPlan(t, dir)
	assert.Contains(t, plan, "untouched body")
	assert.Contains(t, plan, "task-other")
	assert.Contains(t, plan, "slice-b")
}

func TestEnrichTask_EmptyBody(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)
	err := commands.EnrichTask(dir, "slice-a", "task-todo", "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")
}

func TestEnrichTask_InjectionGuard_ID(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)
	err := commands.EnrichTask(dir, "slice-a", "task-todo", "some text\nid: new-id\nmore text")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `must not contain "id:"`)
}

func TestEnrichTask_InjectionGuard_Status(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)
	err := commands.EnrichTask(dir, "slice-a", "task-todo", "status: done")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `must not contain "status:"`)
}

func TestEnrichTask_BlocksInProgress(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)
	err := commands.EnrichTask(dir, "slice-a", "task-in-progress", "new body")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "enrichment skipped")
	assert.Contains(t, err.Error(), "in-progress")
}

func TestEnrichTask_BlocksDone(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)
	err := commands.EnrichTask(dir, "slice-a", "task-done", "new body")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "enrichment skipped")
	assert.Contains(t, err.Error(), "done")
}

func TestEnrichTask_SliceNotFound(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)
	err := commands.EnrichTask(dir, "nonexistent-slice", "task-todo", "new body")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-slice")
}

func TestEnrichTask_TaskNotFound(t *testing.T) {
	dir := t.TempDir()
	writeEnrichPlan(t, dir)
	err := commands.EnrichTask(dir, "slice-a", "nonexistent-task", "new body")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-task")
}

func TestEnrichTask_MissingPlan(t *testing.T) {
	err := commands.EnrichTask(t.TempDir(), "slice-a", "task-todo", "new body")
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "plan.yml")
}
