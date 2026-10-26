package migrator

import (
	"context"
	"database/sql"
	"errors"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) reconcileSuccess(
	ctx context.Context,
	connection *sql.Conn,
	runID string,
	cause error,
) error {
	reconcileContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), runner.cleanupTimeout,
	)
	defer cancel()
	run, runErr := postgresstore.GetMigrationRun(reconcileContext, connection, runID)
	contract, contractErr := postgresstore.GetCoreSchemaContract(reconcileContext, connection)
	if runErr != nil || contractErr != nil {
		return ambiguousOutcome(
			runID, "finalize", errors.Join(cause, runErr, contractErr),
		)
	}
	if runner.validateRunRecord(run) == nil && run.State == "succeeded" &&
		runner.verifyCoreSchemaContract(reconcileContext, connection, contract) == nil {
		return nil
	}
	return ambiguousOutcome(runID, "finalize", cause)
}
