package initialize

import (
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
)

var errGooseNotInstalled = errors.New("goose runner is not installed")

type StageError struct {
	Stage     string
	Phase     string
	Database  string
	Retryable bool
	Code      int
	Secrets   []string
	Err       error
}

func (e *StageError) Error() string {
	cause := ""
	if e.Err != nil {
		cause = foundation.Redact(e.Err.Error(), e.Secrets)
	}
	phase := e.Phase
	if phase == "" {
		phase = e.Stage
	}
	return fmt.Sprintf(
		"initialize stage=%s phase=%s database=%s retryable=%t: %s",
		e.Stage, phase, e.Database, e.Retryable, cause,
	)
}

func (e *StageError) Unwrap() error { return e.Err }

func (e *StageError) ExitCode() int {
	if e.Code == 0 {
		return 1
	}
	return e.Code
}

func stageError(
	stage, phase string, cfg foundation.Config, retryable bool, code int, err error,
) error {
	var already *StageError
	if errors.As(err, &already) {
		return err
	}
	return &StageError{
		Stage: stage, Phase: phase, Database: cfg.Database,
		Retryable: retryable, Code: code, Secrets: cfg.Secrets(), Err: err,
	}
}

func retryableCause(err error) bool {
	var foundationErr *foundation.Error
	if errors.As(err, &foundationErr) {
		return foundationErr.Retryable
	}
	return false
}
