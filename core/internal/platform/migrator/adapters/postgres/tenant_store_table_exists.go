package postgres

import (
	"context"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *TenantStore) TableExists(
	ctx context.Context,
	database goosedb.DBTxConn,
) (bool, error) {
	if err := VerifyTenantMigrationSession(
		ctx, database, store.session, store.target,
	); err != nil {
		return false, err
	}
	var exists bool
	if err := database.QueryRowContext(ctx, tenantCheckpointTableExistsSQL).Scan(&exists); err != nil {
		return false, fmt.Errorf("check tenant checkpoint table: %w", err)
	}
	return exists, nil
}
