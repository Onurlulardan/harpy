package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) finalizeSuccess(
	ctx context.Context,
	connection *sql.Conn,
	runID string,
) error {
	finalContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), runner.cleanupTimeout,
	)
	defer cancel()
	transaction, err := connection.BeginTx(finalContext, nil)
	if err != nil {
		return fmt.Errorf("begin foundation finalization: %w", err)
	}
	defer transaction.Rollback()
	if err := runner.verifyFinalizationTenantSet(finalContext, transaction); err != nil {
		return err
	}
	if err := runner.finalizeRelease(
		finalContext,
		func(
			queryContext context.Context,
			query string,
			arguments ...any,
		) FinalizationRow {
			return transaction.QueryRowContext(queryContext, query, arguments...)
		},
	); err != nil {
		return err
	}
	if err := runner.persistEnvironmentIdentity(finalContext, transaction); err != nil {
		return err
	}
	if err := runner.persistReceiptEvidence(finalContext, transaction); err != nil {
		return err
	}
	if err := runner.persistExtensionEvidence(finalContext, transaction); err != nil {
		return err
	}
	if err := postgresstore.CompleteMigrationRun(finalContext, transaction, runID); err != nil {
		return err
	}
	if err := runner.persistDatabaseGenerationManifest(finalContext, transaction); err != nil {
		return err
	}
	if err := runner.persistReleaseGeneration(finalContext, transaction, runID); err != nil {
		return err
	}
	if err := runner.persistTenantRequirementCatalog(finalContext, transaction, runID); err != nil {
		return err
	}
	contract := runner.coreSchemaContract(runID)
	stored, err := postgresstore.TransitionCoreSchemaContract(
		finalContext, transaction, contract,
	)
	if err != nil {
		return err
	}
	if err := runner.verifyCoreSchemaContract(finalContext, transaction, stored); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit foundation finalization: %w", err)
	}
	return nil
}

func (runner *Runner) coreSchemaContract(
	runID string,
) postgresstore.CoreSchemaContractRecord {
	binaryContract := runner.schemaAuthority.Contract()
	return postgresstore.CoreSchemaContractRecord{
		DatabaseGeneration:       runner.identity.databaseGeneration,
		MinimumRuntimeGeneration: runner.identity.minimumRuntimeGeneration,
		MigrationSHA256:          binaryContract.MigrationManifestSHA256,
		ExtensionSHA256:          binaryContract.ExtensionManifestSHA256,
		Status:                   "ready", VerifiedRunID: runID,
	}
}
