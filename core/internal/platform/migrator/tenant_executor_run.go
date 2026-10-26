package migrator

import (
	"context"
	"errors"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	tenantexec "github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
)

func (executor *TenantExecutor) startRun(
	ctx context.Context, claim tenantexec.Claim,
) error {
	transaction, err := executor.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tenant migration run: %w", err)
	}
	defer transaction.Rollback()
	if claim.PreviousRunID.String() != "00000000-0000-0000-0000-000000000000" {
		_ = postgresstore.FailMigrationRun(
			ctx, transaction, claim.PreviousRunID.String(), "lease_expired",
		)
	}
	if err := postgresstore.CreateMigrationRun(
		ctx, transaction, executor.runInput(claim),
	); err != nil {
		return err
	}
	record, err := postgresstore.GetMigrationRun(
		ctx, transaction, claim.MigrationRunID.String(),
	)
	if err != nil {
		return err
	}
	if err := executor.validateRun(record, claim); err != nil {
		return err
	}
	if record.State == "planned" {
		if err := postgresstore.StartMigrationRun(
			ctx, transaction, claim.MigrationRunID.String(),
		); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return executor.reconcileRun(ctx, claim, "running", err)
	}
	return nil
}

func (executor *TenantExecutor) completeRun(
	ctx context.Context, claim tenantexec.Claim,
) error {
	transaction, err := executor.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if err := postgresstore.CompleteMigrationRun(
		ctx, transaction, claim.MigrationRunID.String(),
	); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return executor.reconcileRun(ctx, claim, "succeeded", err)
	}
	return nil
}

func (executor *TenantExecutor) failRun(
	ctx context.Context, claim tenantexec.Claim, code string,
) {
	_ = postgresstore.FailMigrationRun(
		ctx, executor.database, claim.MigrationRunID.String(), code,
	)
}

func (executor *TenantExecutor) reconcileRun(
	ctx context.Context, claim tenantexec.Claim, expected string, cause error,
) error {
	record, err := postgresstore.GetMigrationRun(
		ctx, executor.database, claim.MigrationRunID.String(),
	)
	if err != nil || executor.validateRun(record, claim) != nil {
		return errors.Join(tenantexec.ErrAmbiguousOutcome, cause, err)
	}
	if record.State != expected {
		return errors.Join(tenantexec.ErrAmbiguousOutcome, cause)
	}
	return nil
}
