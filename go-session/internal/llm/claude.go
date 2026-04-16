package llm

import (
	"fmt"
	"io"
	"os/exec"
)

// claudeRunner implements Runner using `claude --dangerously-skip-permissions`.
type claudeRunner struct{}

func (r *claudeRunner) Run(stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command("claude", "-p", "--dangerously-skip-permissions", `--mcp-config`, `{"mcpServers":{}}`, "--no-session-persistence", "--allowed-tools", "Bash,Read,Glob,Grep,Write", "--model", "haiku")
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude pipeline failed: %w", err)
	}
	return nil
}
