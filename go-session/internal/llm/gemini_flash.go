package llm

import (
	"fmt"
	"io"
	"os/exec"
)

// geminiFlashRunner implements Runner using `gemini --yolo --model=gemini-2.5-flash`.
type geminiFlashRunner struct{}

func (r *geminiFlashRunner) Run(stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command("gemini", "--yolo", "--model=gemini-2.5-flash")
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gemini-flash pipeline failed: %w", err)
	}
	return nil
}
