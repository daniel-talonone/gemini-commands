package gemini

import (
	"fmt"
	"io"
	"os/exec"
)

// RunYolo executes `gemini --yolo` with the given stdin.
func RunYolo(stdin io.Reader, stdout, stderr io.Writer) error {
	geminiCmd := exec.Command("gemini", "--yolo")
	geminiCmd.Stdin = stdin
	geminiCmd.Stdout = stdout
	geminiCmd.Stderr = stderr
	if err := geminiCmd.Run(); err != nil {
		return fmt.Errorf("gemini pipeline failed: %w", err)
	}
	return nil
}
