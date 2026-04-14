package llm

import (
	"fmt"
	"io"
	"os/exec"
)

// geminiRunner implements Runner using `gemini --yolo`.
type geminiRunner struct{}

func (r *geminiRunner) Run(stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command("gemini", "--yolo")
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gemini pipeline failed: %w", err)
	}
	return nil
}

// RunYolo executes `gemini --yolo` with the given stdin.
// Deprecated: prefer NewRunner(ModelGemini).Run(...).
func RunYolo(stdin io.Reader, stdout, stderr io.Writer) error {
	return (&geminiRunner{}).Run(stdin, stdout, stderr)
}
