package llm

import (
	"fmt"
	"io"
)

// Runner abstracts the LLM backend used to execute prompts.
type Runner interface {
	Run(stdin io.Reader, stdout, stderr io.Writer) error
}

// Model selects the LLM backend.
type Model string

const (
	ModelGemini      Model = "gemini"
	ModelGeminiFlash Model = "gemini-flash"
	ModelClaude      Model = "claude"
)

// NewRunner returns a Runner for the given model.
func NewRunner(model Model) (Runner, error) {
	switch model {
	case ModelGemini:
		return &geminiRunner{}, nil
	case ModelGeminiFlash:
		return &geminiFlashRunner{}, nil
	case ModelClaude:
		return &claudeRunner{}, nil
	default:
		return nil, fmt.Errorf("unknown model %q — must be one of: gemini, gemini-flash, claude", model)
	}
}
