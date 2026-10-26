package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func AcquireTenantMigrationLock(
	ctx context.Context,
	database *sql.DB,
	tenantID string,
) error {
	var lockedTenant string
	if err := database.QueryRowContext(
		ctx, acquireTenantMigrationLockSQL, tenantID,
	).Scan(&lockedTenant); err != nil {
		return fmt.Errorf("acquire tenant migration lock: %w", err)
	}
	if lockedTenant != tenantID {
		return fmt.Errorf("tenant migration lock identity mismatch")
	}
	return nil
}
