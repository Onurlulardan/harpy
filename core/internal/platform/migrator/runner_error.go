package migrator

import (
	"errors"
	"fmt"
)

var (
	ErrMigrationsPending = errors.New("platform migrations are pending")
	ErrOutOfOrder        = errors.New("out-of-order platform history is forbidden")
	ErrOutcomeAmbiguous  = errors.New("migration database outcome is ambiguous")
	ErrRunTerminal       = errors.New("migration run is terminal")
	ErrCommitted         = errors.New("migration committed before caller cancellation")
)

type OutcomeError struct {
	RunID string
	Stage string
	Cause error
}

func (failure *OutcomeError) Error() string {
	return fmt.Sprintf("migration run %s %s: %v", failure.RunID, failure.Stage, failure.Cause)
}

func (failure *OutcomeError) Unwrap() error { return failure.Cause }

func ambiguousOutcome(runID, stage string, cause error) error {
	return &OutcomeError{
		RunID: runID, Stage: stage,
		Cause: errors.Join(ErrOutcomeAmbiguous, cause),
	}
}
