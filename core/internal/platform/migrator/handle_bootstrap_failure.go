package migrator

import (
	"context"
	"database/sql"
	"errors"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) handleBootstrapFailure(
	ctx context.Context,
	connection *sql.Conn,
	runID string,
	cause error,
) error {
	reconcileContext, cancel := runner.reconcileContext(ctx)
	defer cancel()
	history, _, historyErr := readHistory(reconcileContext, connection)
	if historyErr != nil {
		return ambiguousOutcome(runID, "bootstrap", errors.Join(cause, historyErr))
	}
	if _, _, historyErr = classifyHistory(
		history, runner.selected, runner.historyAliases,
	); historyErr != nil {
		return historyErr
	}
	ready, readyErr := postgresstore.MigrationRunInfrastructureReady(
		reconcileContext, connection,
	)
	if readyErr != nil {
		return ambiguousOutcome(runID, "bootstrap", errors.Join(cause, readyErr))
	}
	if !ready {
		return &OutcomeError{RunID: runID, Stage: "bootstrap", Cause: cause}
	}
	run, startErr := runner.ensureRunRunning(reconcileContext, connection, runID)
	if startErr != nil {
		run, startErr = runner.reconcileRunStart(
			reconcileContext, connection, runID, startErr,
		)
	}
	if startErr != nil {
		return errors.Join(cause, startErr)
	}
	if run.state == "succeeded" {
		return &OutcomeError{
			RunID: runID, Stage: "bootstrap",
			Cause: errors.Join(ErrCommitted, cause),
		}
	}
	code := "migration_execution_failed"
	if ctx.Err() != nil {
		code = "migration_cancelled"
	}
	return runner.finalizeFailure(ctx, connection, runID, code, cause)
}
