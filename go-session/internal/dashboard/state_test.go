package dashboard_test

import (
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/dashboard"
	"github.com/stretchr/testify/assert"
)

var alwaysAlive = func(int) bool { return true }
var neverAlive = func(int) bool { return false }

func TestDeriveState_Running(t *testing.T) {
	status := &dashboard.FeatureStatus{PID: 999999, PipelineStep: "implement", Mode: "auto", Repo: "org/repo"}
	plan := []dashboard.PlanSlice{
		{ID: "s1", Tasks: []dashboard.PlanTask{{ID: "t1", Status: "done"}}},
	}
	result := dashboard.DeriveState("sc-1", "org/repo", status, plan, alwaysAlive)
	assert.True(t, result.IsRunning)
	assert.Equal(t, "sc-1", result.StoryID)
	assert.Equal(t, "implement", result.PipelineStep)
}

func TestDeriveState_NotRunning(t *testing.T) {
	status := &dashboard.FeatureStatus{PID: 999999}
	result := dashboard.DeriveState("sc-1", "org/repo", status, nil, neverAlive)
	assert.False(t, result.IsRunning)
}

func TestDeriveState_ZeroPIDNotRunning(t *testing.T) {
	status := &dashboard.FeatureStatus{PID: 0}
	result := dashboard.DeriveState("sc-1", "org/repo", status, nil, alwaysAlive)
	assert.False(t, result.IsRunning, "pid=0 must never be considered running")
}

func TestDeriveState_AllDone(t *testing.T) {
	plan := []dashboard.PlanSlice{
		{Tasks: []dashboard.PlanTask{{ID: "t1", Status: "done"}, {ID: "t2", Status: "done"}}},
		{Tasks: []dashboard.PlanTask{{ID: "t3", Status: "done"}}},
	}
	result := dashboard.DeriveState("sc-1", "org/repo", nil, plan, neverAlive)
	assert.True(t, result.AllDone)
}

func TestDeriveState_NotAllDone(t *testing.T) {
	plan := []dashboard.PlanSlice{
		{Tasks: []dashboard.PlanTask{{ID: "t1", Status: "done"}, {ID: "t2", Status: "todo"}}},
	}
	result := dashboard.DeriveState("sc-1", "org/repo", nil, plan, neverAlive)
	assert.False(t, result.AllDone)
}

func TestDeriveState_LastDoneTask(t *testing.T) {
	plan := []dashboard.PlanSlice{
		{Tasks: []dashboard.PlanTask{{ID: "task-a", Status: "done"}, {ID: "task-b", Status: "todo"}}},
		{Tasks: []dashboard.PlanTask{{ID: "task-c", Status: "done"}}},
	}
	result := dashboard.DeriveState("sc-1", "org/repo", nil, plan, neverAlive)
	assert.Equal(t, "task-c", result.LastDoneTask)
}

func TestDeriveState_NoTasks(t *testing.T) {
	result := dashboard.DeriveState("sc-1", "org/repo", nil, nil, neverAlive)
	assert.Equal(t, "", result.LastDoneTask)
	assert.False(t, result.AllDone)
}

func TestDeriveState_NilStatus(t *testing.T) {
	result := dashboard.DeriveState("sc-1", "org/repo", nil, nil, alwaysAlive)
	assert.False(t, result.HasStatus)
	assert.False(t, result.IsRunning)
	assert.Equal(t, "", result.PipelineStep)
	assert.Equal(t, "", result.Mode)
}

func TestDeriveState_RepoFromStatus(t *testing.T) {
	status := &dashboard.FeatureStatus{Repo: "from-status/repo"}
	result := dashboard.DeriveState("sc-1", "path/repo", status, nil, neverAlive)
	assert.Equal(t, "from-status/repo", result.Repo)
}

func TestDeriveState_RepoFallsBackToPath(t *testing.T) {
	status := &dashboard.FeatureStatus{Repo: ""}
	result := dashboard.DeriveState("sc-1", "path/repo", status, nil, neverAlive)
	assert.Equal(t, "path/repo", result.Repo)
}

func TestDeriveState_AllDoneRequiresAtLeastOneTask(t *testing.T) {
	// Plan with slices that have no tasks must not report AllDone=true
	plan := []dashboard.PlanSlice{
		{ID: "s1", Tasks: []dashboard.PlanTask{}},
		{ID: "s2", Tasks: []dashboard.PlanTask{}},
	}
	result := dashboard.DeriveState("sc-1", "org/repo", nil, plan, neverAlive)
	assert.False(t, result.AllDone, "plan with zero tasks across all slices must not be AllDone")
}

func TestDeriveState_AllDoneIgnoresEmptySliceWhenOthersDone(t *testing.T) {
	// An empty-task slice alongside done tasks: all existing tasks are done → AllDone=true
	plan := []dashboard.PlanSlice{
		{ID: "s1", Tasks: []dashboard.PlanTask{{ID: "t1", Status: "done"}}},
		{ID: "s2", Tasks: []dashboard.PlanTask{}},
	}
	result := dashboard.DeriveState("sc-1", "org/repo", nil, plan, neverAlive)
	assert.True(t, result.AllDone, "done tasks + empty slice = all existing tasks done")
}
