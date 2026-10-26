package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func ReleaseTenantMigrationLock(
	ctx context.Context,
	database *sql.DB,
	tenantID string,
) error {
	var releasedTenant string
	var released bool
	if err := database.QueryRowContext(
		ctx, releaseTenantMigrationLockSQL, tenantID,
	).Scan(&releasedTenant, &released); err != nil {
		return fmt.Errorf("release tenant migration lock: %w", err)
	}
	if releasedTenant != tenantID || !released {
		return fmt.Errorf("tenant migration lock was not released")
	}
	return nil
}
