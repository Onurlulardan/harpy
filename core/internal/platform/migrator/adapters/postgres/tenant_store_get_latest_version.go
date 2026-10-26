package postgres

import (
	"context"
	"database/sql"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *TenantStore) GetLatestVersion(
	ctx context.Context,
	database goosedb.DBTxConn,
) (int64, error) {
	var version sql.NullInt64
	if err := database.QueryRowContext(
		ctx, getLatestTenantStoreVersionSQL, store.target.TenantID,
	).Scan(&version); err != nil {
		return -1, fmt.Errorf("get latest tenant migration version: %w", err)
	}
	if !version.Valid {
		return tenantHistoryLatestVersion(version), nil
	}
	return tenantHistoryLatestVersion(version), nil
}

func tenantHistoryLatestVersion(version sql.NullInt64) int64 {
	if !version.Valid {
		return 0
	}
	return version.Int64
}
