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

// splitTestPlan includes big-task-readme in slice-y so we can test ID collision:
// splitting "big-task" with suffix "readme" generates "big-task-readme" which already exists.
const splitTestPlan = `- id: slice-x
  description: First slice
  status: todo
  tasks:
    - id: big-task
      task: do many things
      status: todo
    - id: existing-task
      task: do something else
      status: todo
    - id: task-in-progress
      task: wip
      status: in-progress
    - id: task-done
      task: done
      status: done
- id: slice-y
  description: Second slice
  status: todo
  tasks:
    - id: other-task
      task: untouched
      status: todo
    - id: big-task-readme
      task: this id will collide with big-task + suffix readme
      status: todo
`

func writeSplitPlan(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.yml"), []byte(splitTestPlan), 0644))
}

func readSplitPlan(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "plan.yml"))
	require.NoError(t, err)
	return string(b)
}

func TestSplitTask_ReplacesOriginal(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	require.NoError(t, commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "part-one", Task: "do part one"},
		{Suffix: "part-two", Task: "do part two"},
	}))

	plan := readSplitPlan(t, dir)
	assert.NotContains(t, plan, "id: big-task\n") // original gone
	assert.Contains(t, plan, "big-task-part-one")
	assert.Contains(t, plan, "big-task-part-two")
	assert.Contains(t, plan, "do part one")
	assert.Contains(t, plan, "do part two")
}

func TestSplitTask_GeneratesIDs(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	require.NoError(t, commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "alpha", Task: "alpha task"},
		{Suffix: "beta", Task: "beta task"},
	}))

	plan := readSplitPlan(t, dir)
	assert.Contains(t, plan, "id: big-task-alpha")
	assert.Contains(t, plan, "id: big-task-beta")
}

func TestSplitTask_NewTasksAreTodo(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	require.NoError(t, commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "task a"},
		{Suffix: "b", Task: "task b"},
	}))

	plan := readSplitPlan(t, dir)
	// Both new tasks must be todo
	assert.Contains(t, plan, "big-task-a")
	assert.Contains(t, plan, "big-task-b")
	// Ensure status todo appears for them (fragile but sufficient)
	assert.Contains(t, plan, "status: todo")
}

func TestSplitTask_OtherTasksUntouched(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	require.NoError(t, commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "task a"},
		{Suffix: "b", Task: "task b"},
	}))

	plan := readSplitPlan(t, dir)
	assert.Contains(t, plan, "other-task")
	assert.Contains(t, plan, "untouched")
	assert.Contains(t, plan, "slice-y")
	assert.Contains(t, plan, "existing-task")
}

func TestSplitTask_TooFewReplacements(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "only", Task: "only one task"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2")
}

func TestSplitTask_EmptyReplacements(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2")
}

func TestSplitTask_InvalidSuffix(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "Bad Suffix", Task: "task a"},
		{Suffix: "good-suffix", Task: "task b"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not kebab-case")
}

func TestSplitTask_InjectionGuard_ID(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "some text\nid: injected\nmore text"},
		{Suffix: "b", Task: "fine"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"id:"`)
}

func TestSplitTask_InjectionGuard_Status(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "status: done"},
		{Suffix: "b", Task: "fine"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"status:"`)
}

func TestSplitTask_BlocksInProgress(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "task-in-progress", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "part a"},
		{Suffix: "b", Task: "part b"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "split skipped")
	assert.Contains(t, err.Error(), "in-progress")
}

func TestSplitTask_BlocksDone(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "task-done", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "part a"},
		{Suffix: "b", Task: "part b"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "split skipped")
	assert.Contains(t, err.Error(), "done")
}

func TestSplitTask_DuplicateSuffix(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "same", Task: "part a"},
		{Suffix: "same", Task: "part b"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collides")
}

func TestSplitTask_CollidesWithExistingTask(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	// "big-task" + suffix "readme" → "big-task-readme" which exists in slice-y
	err := commands.SplitTask(dir, "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "readme", Task: "part a"},
		{Suffix: "agents", Task: "part b"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collides")
	assert.Contains(t, err.Error(), "big-task-readme")
}

func TestSplitTask_SliceNotFound(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "nonexistent-slice", "big-task", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "part a"},
		{Suffix: "b", Task: "part b"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-slice")
}

func TestSplitTask_TaskNotFound(t *testing.T) {
	dir := t.TempDir()
	writeSplitPlan(t, dir)

	err := commands.SplitTask(dir, "slice-x", "nonexistent-task", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "part a"},
		{Suffix: "b", Task: "part b"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-task")
}

func TestSplitTask_MissingPlan(t *testing.T) {
	err := commands.SplitTask(t.TempDir(), "slice-x", "big-task", []commands.SplitTaskEntry{
		{Suffix: "a", Task: "part a"},
		{Suffix: "b", Task: "part b"},
	})
	require.Error(t, err)
	assert.True(t, strings.Contains(strings.ToLower(err.Error()), "plan.yml"))
}
