package migrator

import (
	"context"
	"errors"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/jackc/pgx/v5"
)

func (runner *Runner) executeTenantTarget(
	ctx context.Context,
	target postgresstore.TenantMigrationTarget,
	runID string,
) (_ []int64, resultErr error) {
	database, err := runner.tenantDatabase()
	if err != nil {
		return nil, fmt.Errorf("open tenant migration database: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, database.Close()) }()
	if database.Stats().MaxOpenConnections != 1 {
		return nil, fmt.Errorf("tenant migration database must be one-shot single-session")
	}
	if err := postgresstore.AcquireTenantMigrationLock(ctx, database, target.TenantID); err != nil {
		return nil, err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), requiredCleanupTimeout)
		defer cancel()
		resultErr = errors.Join(
			resultErr,
			postgresstore.ReleaseTenantMigrationLock(cleanup, database, target.TenantID),
		)
	}()
	searchPath := pgx.Identifier{target.SchemaName}.Sanitize() + ",pg_catalog"
	if err := postgresstore.SetTenantMigrationSearchPath(ctx, database, searchPath); err != nil {
		return nil, err
	}
	session, err := postgresstore.ReadTenantMigrationSession(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("bind tenant migration session: %w", err)
	}
	if session.SearchPath != searchPath {
		return nil, fmt.Errorf("tenant migration session search path differs from target")
	}
	if session.SchemaName != target.SchemaName {
		return nil, fmt.Errorf("tenant migration target schema is absent from session search path")
	}
	provider, err := runner.buildTenantGooseProvider(database, target, session, runID)
	if err != nil {
		return nil, err
	}
	var applied []int64
	var committedErr error
	for _, selected := range runner.tenantSelected {
		apply, err := prepareTenantCheckpoint(
			ctx, database, target, session, selected, runner.tenantAliases, runID,
		)
		if err != nil {
			return nil, err
		}
		if !apply {
			continue
		}
		if _, err := provider.UpTo(ctx, selected.Version); err != nil {
			reconcileCtx, cancelReconcile := context.WithTimeout(
				context.WithoutCancel(ctx), requiredCleanupTimeout,
			)
			committed, safelyFailed, reconcileErr := reconcileTenantMigration(
				reconcileCtx, database, target, session, selected,
				runner.tenantAliases,
			)
			cancelReconcile()
			if reconcileErr != nil {
				return nil, ambiguousOutcome(
					runID, "tenant migration reconcile", errors.Join(err, reconcileErr),
				)
			}
			if committed {
				applied = append(applied, selected.Version)
				committedErr = errors.Join(committedErr, ErrCommitted, err)
				continue
			}
			if !safelyFailed {
				return nil, err
			}
			code := "migration_execution_failed"
			if ctx.Err() != nil {
				code = "migration_cancelled"
			}
			failureCtx, cancel := context.WithTimeout(
				context.WithoutCancel(ctx), requiredCleanupTimeout,
			)
			failureErr := failTenantCheckpoint(
				failureCtx, database, target, session, selected, runID, code,
			)
			cancel()
			return nil, errors.Join(err, failureErr)
		}
		if err := verifyTenantCheckpoint(
			ctx, database, target, selected, runner.tenantAliases,
		); err != nil {
			return nil, err
		}
		applied = append(applied, selected.Version)
	}
	if err := runner.verifyTenantIdentity(ctx, database, target, session); err != nil {
		return nil, err
	}
	return applied, committedErr
}
