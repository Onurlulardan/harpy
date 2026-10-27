package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) finalizeFailure(
	ctx context.Context,
	connection *sql.Conn,
	runID, failureCode string,
	cause error,
) error {
	finalContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), runner.cleanupTimeout,
	)
	defer cancel()
	err := postgresstore.FailMigrationRun(finalContext, connection, runID, failureCode)
	if err == nil {
		return cause
	}
	record, readErr := postgresstore.GetMigrationRun(finalContext, connection, runID)
	if readErr != nil {
		return ambiguousOutcome(runID, "fail", errors.Join(cause, err, readErr))
	}
	if record.State == "failed" && record.FailureCode == failureCode {
		return cause
	}
	if record.State == "succeeded" {
		return &OutcomeError{
			RunID: runID, Stage: "fail",
			Cause: errors.Join(ErrCommitted, cause),
		}
	}
	return ambiguousOutcome(
		runID, "fail",
		errors.Join(cause, err, fmt.Errorf("observed run state %s", record.State)),
	)
}
