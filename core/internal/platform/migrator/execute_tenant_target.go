package migrator

import (
	"context"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/baseline"
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
	appliedResult, err := baseline.ApplyTenantDB(ctx, database, target.SchemaName)
	if err != nil {
		return nil, err
	}
	var applied []int64
	if appliedResult.Applied {
		applied = []int64{appliedResult.Version}
	}
	var committedErr error
	if err := runner.verifyTenantIdentity(ctx, database, target, session); err != nil {
		return nil, err
	}
	return applied, committedErr
}
