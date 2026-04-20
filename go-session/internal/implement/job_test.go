package implement

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noopRunner is a Runner that always succeeds without invoking any external process.
// It is only exercised when IN_TEST_MODE is false; job tests run with IN_TEST_MODE=true
// so the runner is never called in practice.
type noopRunner struct{}

func (r *noopRunner) Run(_ io.Reader, _, _ io.Writer) error { return nil }

type mockJob struct {
	prompt         string
	onSuccess      func(attempt int) error
	onFailure      func(attempt int) error
	onSuccessCount int
	onFailureCount int
}

func (m *mockJob) Prompt() (string, error) {
	return m.prompt, nil
}

func (m *mockJob) OnSuccess(attempt int) error {
	m.onSuccessCount++
	if m.onSuccess != nil {
		return m.onSuccess(attempt)
	}
	return nil
}

func (m *mockJob) OnFailure(attempt int) error {
	m.onFailureCount++
	if m.onFailure != nil {
		return m.onFailure(attempt)
	}
	return nil
}

func TestRunJob_SuccessFirstTry(t *testing.T) {
	t.Setenv("IN_TEST_MODE", "true")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureDir := t.TempDir()
	successCalled := false
	job := &mockJob{
		prompt: "test prompt",
		onSuccess: func(attempt int) error {
			successCalled = true
			return nil
		},
		onFailure: func(attempt int) error {
			t.Fatal("onFailure should not be called")
			return nil
		},
	}

	err := RunJob(featureDir, 3, 0, logger, &noopRunner{}, job)
	assert.NoError(t, err)
	assert.True(t, successCalled)
}

func TestRunJob_SuccessOnRetry(t *testing.T) {
	t.Setenv("IN_TEST_MODE", "true")

	tmpfile, err := os.CreateTemp("", "test-retry")
	require.NoError(t, err)
	fileName := tmpfile.Name()
	_ = tmpfile.Close()
	_ = os.Remove(fileName)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureDir := t.TempDir()
	callCount := 0
	job := &mockJob{
		prompt: "test prompt",
		onSuccess: func(attempt int) error {
			callCount++
			if callCount == 1 {
				return fmt.Errorf("not ready yet")
			}
			return nil
		},
	}

	err = RunJob(featureDir, 3, 0, logger, &noopRunner{}, job)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestRunJob_ExhaustedRetries(t *testing.T) {
	t.Setenv("IN_TEST_MODE", "true")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureDir := t.TempDir()
	maxRetries := 3
	job := &mockJob{
		prompt: "test prompt",
		onSuccess: func(attempt int) error {
			return fmt.Errorf("always fails")
		},
	}

	err := RunJob(featureDir, maxRetries, 0, logger, &noopRunner{}, job)

	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("job failed after %d attempts", maxRetries))
	assert.Equal(t, maxRetries, job.onSuccessCount)
}

func TestRunJob_OnSuccessError(t *testing.T) {
	t.Setenv("IN_TEST_MODE", "true")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureDir := t.TempDir()
	expectedErr := fmt.Errorf("onSuccess hook error")

	job := &mockJob{
		prompt: "test prompt",
		onSuccess: func(attempt int) error {
			return expectedErr
		},
		onFailure: func(attempt int) error {
			t.Fatal("onFailure should not be called")
			return nil
		},
	}

	err := RunJob(featureDir, 1, 0, logger, &noopRunner{}, job)
	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}

func TestRunJob_OnFailureError(t *testing.T) {
	t.Setenv("IN_TEST_MODE", "true")

	// OnFailure is only called when Gemini exits non-zero, which can't happen in
	// test mode (Gemini is skipped entirely). This is a known limitation of the
	// IN_TEST_MODE seam — testing this path requires injecting a Gemini error,
	// which requires dependency injection (a separate refactor).
	t.Skip("OnFailure path requires Gemini dependency injection to test")
}
