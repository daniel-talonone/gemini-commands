package implement

// Job defines the interface for a runnable unit of work in the implement pipeline.
type Job interface {
	// Prompt assembles and returns the instruction to send to the AI.
	// It may perform I/O (e.g. reading plan.yml, template files) on every call.
	// Returns an error if the prompt cannot be assembled.
	Prompt() (string, error)
	// OnSuccess is called when the AI process exits successfully. It should validate
	// the result (e.g. check task statuses, run verification) and return nil only when
	// the job is complete. A non-nil return triggers a retry.
	OnSuccess(attempt int) error
	// OnFailure is called when the AI process exits with a non-zero code (rate-limit
	// errors are retried transparently and never reach this hook). It should log the
	// failure context and return an error to trigger a retry, or nil if the job can
	// be considered complete despite the Gemini error.
	OnFailure(attempt int) error
}
