package implement_test

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/implement"
	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRunner is a no-op Runner for use in tests (IN_TEST_MODE bypasses the
// actual Run call, so this is never invoked).
type testRunner struct{}

func (r *testRunner) Run(_ io.Reader, _, _ io.Writer) error { return nil }

// perSliceTestEnv holds the directories created by setupPerSliceTest.
// verifyScript is the path to the verification shell script; overwrite its
// content before calling implement.Run to change verification behaviour.
type perSliceTestEnv struct {
	featureDir    string
	projectDir    string
	aiSessionHome string
	verifyScript  string
	logger        *slog.Logger
}

// setupPerSliceTest creates a minimal test environment for PerSliceStrategy
// tests. The verification script defaults to always-success (exit 0).
func setupPerSliceTest(t *testing.T, planYML string) perSliceTestEnv {
	t.Helper()
	t.Setenv("IN_TEST_MODE", "true")

	// Verification script — tests can overwrite the content to change behaviour.
	scriptsDir := t.TempDir()
	verifyScript := filepath.Join(scriptsDir, "verify.sh")
	require.NoError(t, os.WriteFile(verifyScript, []byte(`#!/bin/bash
exit 0`), 0755))

	// projectDir holds AGENTS.md which points at the verify script.
	projectDir := t.TempDir()
	agentsMD := "## Verification\nRun: " + verifyScript + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "AGENTS.md"), []byte(agentsMD), 0644))

	// aiSessionHome needs execute_slice.md with all placeholder tokens.
	aiSessionHome := t.TempDir()
	headlessDir := filepath.Join(aiSessionHome, "headless", "session")
	require.NoError(t, os.MkdirAll(headlessDir, 0755))
	sliceMD := "{{story_description_here}} {{architecture_description_here}} {{slice_description_here}} {{tasks_here}} {{changes_so_far_here}} {{verification_command_here}} {{feature_dir_here}}"
	require.NoError(t, os.WriteFile(filepath.Join(headlessDir, "execute_slice.md"), []byte(sliceMD), 0644))
	// execute_task.md is required by PerTaskStrategy (used in TestRun).
	require.NoError(t, os.WriteFile(filepath.Join(headlessDir, "execute_task.md"), []byte("prompt"), 0644))

	// featureDir holds the feature files.
	featureDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "description.md"), []byte("Test story"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "architecture.md"), []byte("Test arch"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "status.yaml"), []byte("pipeline_step: plan-done\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "plan.yml"), []byte(planYML), 0644))

	return perSliceTestEnv{
		featureDir:    featureDir,
		projectDir:    projectDir,
		aiSessionHome: aiSessionHome,
		verifyScript:  verifyScript,
		logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func TestRun(t *testing.T) {
	t.Setenv("IN_TEST_MODE", "true")

	planYML := `- id: test-slice
  description: "Test slice"
  status: todo
  depends_on: []
  tasks:
    - id: test-task
      task: "Do something"
      status: todo
    - id: done-task
      task: "Already done"
      status: done
`
	env := setupPerSliceTest(t, planYML)

	err := implement.Run(env.logger, "test-feature", env.featureDir, env.projectDir, env.aiSessionHome, 3, 0, &implement.PerTaskStrategy{}, &testRunner{})
	require.NoError(t, err)

	// plan.yml: test-task and test-slice must be marked done.
	planBytes, err := os.ReadFile(filepath.Join(env.featureDir, "plan.yml"))
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(planBytes), "status: done"))

	// log.md must record completion.
	logBytes, err := os.ReadFile(filepath.Join(env.featureDir, "log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(logBytes), "IMPLEMENT COMPLETE")

	// status.yaml must be updated to implement-done.
	statusBytes, err := os.ReadFile(filepath.Join(env.featureDir, "status.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(statusBytes), "implement-done")
}

func TestExtractVerificationCommand(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		tempDir := t.TempDir()
		agentsMD := `## Verification
Run: echo 'hello world'
`
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "AGENTS.md"), []byte(agentsMD), 0644))

		cmd, err := implement.ExtractVerificationCommand(tempDir)
		require.NoError(t, err)
		assert.Equal(t, "echo 'hello world'", cmd)
	})

	t.Run("missing AGENTS.md", func(t *testing.T) {
		_, err := implement.ExtractVerificationCommand(t.TempDir())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading AGENTS.md")
	})

	t.Run("missing verification section", func(t *testing.T) {
		tempDir := t.TempDir()
		agentsMD := `## Some Other Section
Content here.
`
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "AGENTS.md"), []byte(agentsMD), 0644))

		_, err := implement.ExtractVerificationCommand(tempDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "verification command not found")
	})
}

func TestPerSliceStrategy_ExecuteSlice_Success(t *testing.T) {
	planYML := `- id: test-slice-success
  description: "Test slice success"
  status: todo
  depends_on: []
  tasks:
    - id: task-1
      task: "Complete this task"
      status: todo
`
	env := setupPerSliceTest(t, planYML)

	// Simulate Gemini marking the task done (IN_TEST_MODE skips the real call).
	require.NoError(t, plan.UpdateTask(env.featureDir, "task-1", "done"))

	err := implement.Run(env.logger, "test-feature-success", env.featureDir, env.projectDir, env.aiSessionHome, 1, 0, &implement.PerSliceStrategy{}, &testRunner{})
	require.NoError(t, err)

	// Slice and task must be marked done.
	p, err := plan.LoadPlan(env.featureDir)
	require.NoError(t, err)
	slice, found := p.FindSlice("test-slice-success")
	require.True(t, found)
	assert.Equal(t, "done", slice.Status)
	assert.Equal(t, "done", slice.Tasks[0].Status)

	logBytes, err := os.ReadFile(filepath.Join(env.featureDir, "log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(logBytes), "Slice test-slice-success: all gates passed (attempt 1).")
	assert.Contains(t, string(logBytes), "--- IMPLEMENT COMPLETE ---")

	statusBytes, err := os.ReadFile(filepath.Join(env.featureDir, "status.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(statusBytes), "implement-done")
}

func TestPerSliceStrategy_ExecuteSlice_Gate1RetryFails(t *testing.T) {
	planYML := `- id: test-slice-gate1-fail
  description: "Test slice Gate 1 fail"
  status: todo
  depends_on: []
  tasks:
    - id: task-g1-1
      task: "Complete this task"
      status: todo
`
	env := setupPerSliceTest(t, planYML)
	// Tasks never get marked done — Gate 1 fails every attempt.

	err := implement.Run(env.logger, "test-feature-gate1-fail", env.featureDir, env.projectDir, env.aiSessionHome, 2, 0, &implement.PerSliceStrategy{}, &testRunner{})
	require.Error(t, err)

	p, err := plan.LoadPlan(env.featureDir)
	require.NoError(t, err)
	slice, found := p.FindSlice("test-slice-gate1-fail")
	require.True(t, found)
	assert.Equal(t, "in-progress", slice.Status)
	assert.Equal(t, "todo", slice.Tasks[0].Status)

	logBytes, err := os.ReadFile(filepath.Join(env.featureDir, "log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(logBytes), "Gate 1 failed for slice test-slice-gate1-fail (attempt 1)")
	assert.Contains(t, string(logBytes), "Gate 1 failed for slice test-slice-gate1-fail (attempt 2)")

	statusBytes, err := os.ReadFile(filepath.Join(env.featureDir, "status.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(statusBytes), "implement-failed")
}

func TestPerSliceStrategy_ExecuteSlice_Gate2RetrySucceeds(t *testing.T) {
	planYML := `- id: test-slice-gate2-retry
  description: "Test slice Gate 2 retry"
  status: todo
  depends_on: []
  tasks:
    - id: task-g2-1
      task: "Complete this task"
      status: todo
`
	env := setupPerSliceTest(t, planYML)

	// Tasks pre-marked done so Gate 1 always passes.
	require.NoError(t, plan.UpdateTask(env.featureDir, "task-g2-1", "done"))

	// Script call sequence: initial gate (pass), Gate 2 attempt 1 (fail), Gate 2 attempt 2 (pass).
	counterFile := filepath.Join(t.TempDir(), "count")
	require.NoError(t, os.WriteFile(env.verifyScript, []byte(`#!/bin/bash
COUNT=$(cat "`+counterFile+`" 2>/dev/null || echo 0)
COUNT=$((COUNT + 1))
echo "$COUNT" > "`+counterFile+`"
[ "$COUNT" -le 1 ] && exit 0   # initial gate: pass
[ "$COUNT" -le 2 ] && exit 1   # Gate 2 attempt 1: fail
exit 0                          # Gate 2 attempt 2: pass`), 0755))

	err := implement.Run(env.logger, "test-feature-gate2-retry", env.featureDir, env.projectDir, env.aiSessionHome, 3, 0, &implement.PerSliceStrategy{}, &testRunner{})
	require.NoError(t, err)

	logBytes, err := os.ReadFile(filepath.Join(env.featureDir, "log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(logBytes), "Gate 2 failed for slice test-slice-gate2-retry (attempt 1)")
	assert.Contains(t, string(logBytes), "Slice test-slice-gate2-retry: all gates passed (attempt 2).")
}

func TestPerSliceStrategy_ExecuteSlice_Gate2ExhaustedRetries(t *testing.T) {
	planYML := `- id: test-slice-gate2-exhausted
  description: "Test slice Gate 2 exhausted"
  status: todo
  depends_on: []
  tasks:
    - id: task-g2ex-1
      task: "Complete this task"
      status: todo
`
	env := setupPerSliceTest(t, planYML)

	// Tasks pre-marked done so Gate 1 always passes.
	// Script: initial gate passes, all Gate 2 attempts fail.
	require.NoError(t, plan.UpdateTask(env.featureDir, "task-g2ex-1", "done"))
	counterFile := filepath.Join(t.TempDir(), "count")
	require.NoError(t, os.WriteFile(env.verifyScript, []byte(`#!/bin/bash
COUNT=$(cat "`+counterFile+`" 2>/dev/null || echo 0)
COUNT=$((COUNT + 1))
echo "$COUNT" > "`+counterFile+`"
[ "$COUNT" -le 1 ] && exit 0  # initial gate: pass
exit 1                         # Gate 2: always fail`), 0755))

	err := implement.Run(env.logger, "test-feature-gate2-exhausted", env.featureDir, env.projectDir, env.aiSessionHome, 2, 0, &implement.PerSliceStrategy{}, &testRunner{})
	require.Error(t, err)

	logBytes, err := os.ReadFile(filepath.Join(env.featureDir, "log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(logBytes), "Gate 2 failed for slice test-slice-gate2-exhausted (attempt 1)")
	assert.Contains(t, string(logBytes), "Gate 2 failed for slice test-slice-gate2-exhausted (attempt 2)")

	statusBytes, err := os.ReadFile(filepath.Join(env.featureDir, "status.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(statusBytes), "implement-failed")
}

func TestImplementRun_InitialVerificationFails(t *testing.T) {
	planYML := `- id: test-slice-initial-fail
  description: "Test slice initial fail"
  status: todo
  depends_on: []
  tasks:
    - id: task-initial-1
      task: "Complete this task"
      status: todo
`
	env := setupPerSliceTest(t, planYML)

	// Make initial verification fail before any slice runs.
	require.NoError(t, os.WriteFile(env.verifyScript, []byte(`#!/bin/bash
exit 1`), 0755))

	err := implement.Run(env.logger, "test-feature-initial-fail", env.featureDir, env.projectDir, env.aiSessionHome, 1, 0, &implement.PerSliceStrategy{}, &testRunner{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Initial verification failed")

	// Slice and task must remain untouched.
	p, err := plan.LoadPlan(env.featureDir)
	require.NoError(t, err)
	slice, found := p.FindSlice("test-slice-initial-fail")
	require.True(t, found)
	assert.Equal(t, "todo", slice.Status)
	assert.Equal(t, "todo", slice.Tasks[0].Status)

	statusBytes, err := os.ReadFile(filepath.Join(env.featureDir, "status.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(statusBytes), "implement-failed")
}
