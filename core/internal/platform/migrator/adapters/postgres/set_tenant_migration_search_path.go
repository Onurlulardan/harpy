package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func SetTenantMigrationSearchPath(
	ctx context.Context,
	database *sql.DB,
	searchPath string,
) error {
	var applied string
	if err := database.QueryRowContext(
		ctx, setTenantMigrationSearchPathSQL, searchPath,
	).Scan(&applied); err != nil {
		return fmt.Errorf("set tenant migration search path: %w", err)
	}
	if applied != searchPath {
		return fmt.Errorf("tenant migration search path mismatch")
	}
	return nil
}
