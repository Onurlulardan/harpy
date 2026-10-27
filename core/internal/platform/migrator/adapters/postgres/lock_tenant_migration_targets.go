package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed sql/lock_tenant_migration_targets.sql
var lockTenantMigrationTargetsSQL string

func LockTenantMigrationTargets(ctx context.Context, transaction *sql.Tx) error {
	if _, err := transaction.ExecContext(ctx, lockTenantMigrationTargetsSQL); err != nil {
		return fmt.Errorf("lock tenant migration targets: %w", err)
	}
	return nil
}
