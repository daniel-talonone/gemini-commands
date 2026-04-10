package implement_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daniel-talonone/gemini-commands/internal/implement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Setenv("IN_TEST_MODE", "true")

	// projectDir acts as the target project root (workDir); contains AGENTS.md.
	projectDir := t.TempDir()
	agentsMD := "## Verification\nRun: echo ok\n"
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "AGENTS.md"), []byte(agentsMD), 0644))

	// aiSessionHome only needs headless/session/execute_task.md to exist.
	aiSessionHome := t.TempDir()
	headlessDir := filepath.Join(aiSessionHome, "headless", "session")
	require.NoError(t, os.MkdirAll(headlessDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(headlessDir, "execute_task.md"), []byte("prompt"), 0644))

	// featureDir contains description.md and plan.yml.
	featureDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "description.md"), []byte("Test story"), 0644))
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
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "plan.yml"), []byte(planYML), 0644))
	// status.yaml must exist for status.Write to update it without error.
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "status.yaml"), []byte("pipeline_step: plan-done\n"), 0644))

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	err := implement.Run(logger, "test-feature", featureDir, projectDir, aiSessionHome, 3, 0, &implement.PerTaskStrategy{})
	require.NoError(t, err)

	// Verify plan.yml: test-task and test-slice must be marked done.
	planBytes, err := os.ReadFile(filepath.Join(featureDir, "plan.yml"))
	require.NoError(t, err)
	planContent := string(planBytes)
	assert.True(t, strings.Contains(planContent, "status: done"), "plan.yml should contain at least one done status")

	// Verify log.md was written.
	logBytes, err := os.ReadFile(filepath.Join(featureDir, "log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(logBytes), "IMPLEMENT COMPLETE")

	// Verify status.yaml was updated to implement-done.
	statusBytes, err := os.ReadFile(filepath.Join(featureDir, "status.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(statusBytes), "implement-done")
}
