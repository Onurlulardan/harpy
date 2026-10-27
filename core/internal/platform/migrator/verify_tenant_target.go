package migrator

import (
	"context"
	"errors"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/jackc/pgx/v5"
)

func (runner *Runner) verifyTenantTarget(
	ctx context.Context,
	target postgresstore.TenantMigrationTarget,
) (resultErr error) {
	database, err := runner.tenantDatabase()
	if err != nil {
		return fmt.Errorf("open tenant verification database: %w", err)
	}
	defer func() {
		joinTenantVerificationDatabaseClose(&resultErr, database)
	}()
	if database.Stats().MaxOpenConnections != 1 {
		return fmt.Errorf("tenant verification database must be single-session")
	}
	if err := postgresstore.AcquireTenantMigrationLock(ctx, database, target.TenantID); err != nil {
		return err
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
		return err
	}
	session, readErr := postgresstore.ReadTenantMigrationSession(ctx, database)
	if readErr != nil {
		return readErr
	}
	if session.SearchPath != searchPath || session.SchemaName != target.SchemaName {
		return fmt.Errorf("tenant verification session differs from target")
	}
	for _, selected := range runner.tenantSelected {
		if err := verifyTenantCheckpoint(
			ctx, database, target, selected, runner.tenantAliases,
		); err != nil {
			return err
		}
	}
	return runner.verifyTenantIdentity(ctx, database, target, session)
}

type tenantVerificationDatabaseCloser interface {
	Close() error
}

func joinTenantVerificationDatabaseClose(
	resultErr *error,
	database tenantVerificationDatabaseCloser,
) {
	closeErr := database.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close tenant verification database: %w", closeErr)
	}
	*resultErr = errors.Join(*resultErr, closeErr)
}
