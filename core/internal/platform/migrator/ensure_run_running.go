package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) ensureRunRunning(
	ctx context.Context,
	connection *sql.Conn,
	runID string,
) (runRecord, error) {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return runRecord{}, fmt.Errorf("begin migration run transaction: %w", err)
	}
	defer transaction.Rollback()
	if err := postgresstore.CreateMigrationRun(
		ctx, transaction, runner.createRunInput(runID),
	); err != nil {
		return runRecord{}, err
	}
	record, err := postgresstore.GetMigrationRun(ctx, transaction, runID)
	if err != nil {
		return runRecord{}, err
	}
	if err := runner.validateRunRecord(record); err != nil {
		return runRecord{}, err
	}
	if record.State == "planned" {
		if err := postgresstore.StartMigrationRun(ctx, transaction, runID); err != nil {
			return runRecord{}, err
		}
		record.State = "running"
	}
	if err := transaction.Commit(); err != nil {
		return runRecord{}, fmt.Errorf("commit migration run start: %w", err)
	}
	return toRunRecord(record), nil
}

func (runner *Runner) reconcileRunStart(
	ctx context.Context,
	connection *sql.Conn,
	runID string,
	cause error,
) (runRecord, error) {
	cleanupContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), runner.cleanupTimeout,
	)
	defer cancel()
	record, err := postgresstore.GetMigrationRun(cleanupContext, connection, runID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return runRecord{}, cause
		}
		return runRecord{}, ambiguousOutcome(runID, "start", errors.Join(cause, err))
	}
	if err := runner.validateRunRecord(record); err != nil {
		return runRecord{}, err
	}
	if record.State != "running" && record.State != "succeeded" {
		return runRecord{}, cause
	}
	return toRunRecord(record), nil
}

func (runner *Runner) createRunInput(runID string) postgresstore.CreateMigrationRunInput {
	identity := runner.identity
	return postgresstore.CreateMigrationRunInput{
		RunID: runID, ReleaseID: identity.releaseID,
		EnvironmentID: identity.environmentID, DatabaseOrigin: string(identity.databaseOrigin),
		Command: "up", MigrationManifestSHA256: identity.manifestSHA256,
		ExtensionManifestSHA256:  identity.extensionSHA256,
		RolesManifestSHA256:      identity.rolesSHA256,
		PrivilegesManifestSHA256: identity.privilegesSHA256,
		OwnershipManifestSHA256:  identity.ownershipSHA256,
	}
}
