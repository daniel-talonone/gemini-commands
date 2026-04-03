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

const validPlanYAML = `- id: slice-one
  description: First slice
  status: todo
  tasks:
    - id: task-one
      task: Do thing 1
      status: todo
    - id: task-two
      task: Do thing 2
      status: in-progress
- id: slice-two
  description: Second slice
  status: done
  tasks:
    - id: task-three
      task: Do thing 3
      status: done
`

func TestValidatePlan_Valid(t *testing.T) {
	require.NoError(t, commands.ValidatePlan([]byte(validPlanYAML)))
}

func TestValidatePlan_Errors(t *testing.T) {
	minTask := func(id string) string {
		return "    - id: " + id + "\n      task: do it\n      status: todo\n"
	}
	minSlice := func(id, desc string, tasks string) string {
		return "- id: " + id + "\n  description: " + desc + "\n  status: todo\n  tasks:\n" + tasks
	}
	valid := minSlice("s", "desc", minTask("t"))

	cases := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"invalid yaml", "not: valid: [yaml", "invalid YAML"},
		{"empty sequence", "[]", "at least one slice"},
		{"empty string", "", "at least one slice"},
		{"slice missing id", "- description: foo\n  status: todo\n  tasks:\n" + minTask("t"), `missing field "id"`},
		{"slice id not kebab", minSlice("Bad ID", "d", minTask("t")), "not kebab-case"},
		{"duplicate slice id", valid + valid, "duplicate slice id"},
		{"slice missing description", "- id: s\n  status: todo\n  tasks:\n" + minTask("t"), `missing field "description"`},
		{"slice invalid status", "- id: s\n  description: d\n  status: nope\n  tasks:\n    - id: t\n      task: do it\n      status: todo\n", "invalid status"},
		{"slice empty tasks", "- id: s\n  description: d\n  status: todo\n  tasks: []\n", "non-empty list"},
		{"task missing id", "- id: s\n  description: d\n  status: todo\n  tasks:\n    - task: t\n      status: todo\n", `missing field "id"`},
		{"task id not kebab", "- id: s\n  description: d\n  status: todo\n  tasks:\n    - id: Bad ID\n      task: t\n      status: todo\n", "not kebab-case"},
		{"duplicate task id cross-slice", minSlice("s1", "d", minTask("dup")) + minSlice("s2", "d", minTask("dup")), "duplicate task id"},
		{"task missing body", "- id: s\n  description: d\n  status: todo\n  tasks:\n    - id: t\n      status: todo\n", `missing field "task"`},
		{"task invalid status", "- id: s\n  description: d\n  status: todo\n  tasks:\n    - id: t\n      task: do it\n      status: nope\n", "invalid status"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := commands.ValidatePlan([]byte(tc.input))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestWritePlan_WritesExactBytes(t *testing.T) {
	dir := t.TempDir()
	data := []byte(validPlanYAML)
	require.NoError(t, commands.WritePlan(dir, data))

	got, err := os.ReadFile(filepath.Join(dir, "plan.yml"))
	require.NoError(t, err)
	assert.Equal(t, string(data), string(got), "bytes must be preserved exactly — no reformatting")
}

func TestWritePlan_InvalidDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	err := commands.WritePlan(dir, []byte("not: valid: yaml"))
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid YAML") || strings.Contains(err.Error(), "at least one slice"))

	_, statErr := os.Stat(filepath.Join(dir, "plan.yml"))
	assert.True(t, os.IsNotExist(statErr), "plan.yml must not be written on validation failure")

	_, tmpErr := os.Stat(filepath.Join(dir, "plan.yml.tmp"))
	assert.True(t, os.IsNotExist(tmpErr), ".tmp file must not linger")
}
