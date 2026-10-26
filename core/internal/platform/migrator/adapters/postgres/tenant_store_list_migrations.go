package postgres

import (
	"context"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *TenantStore) ListMigrations(
	ctx context.Context,
	database goosedb.DBTxConn,
) ([]*goosedb.ListMigrationsResult, error) {
	rows, err := database.QueryContext(
		ctx, listTenantStoreMigrationsSQL, store.target.TenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tenant migration checkpoints: %w", err)
	}
	defer rows.Close()
	migrations := tenantHistoryWithAnchor(nil)
	for rows.Next() {
		var migration goosedb.ListMigrationsResult
		if err := rows.Scan(&migration.Version, &migration.IsApplied); err != nil {
			return nil, fmt.Errorf("scan tenant migration checkpoint: %w", err)
		}
		migrations = append(migrations, &migration)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenant migration checkpoints: %w", err)
	}
	return migrations, nil
}

func tenantHistoryWithAnchor(
	migrations []*goosedb.ListMigrationsResult,
) []*goosedb.ListMigrationsResult {
	return append(
		[]*goosedb.ListMigrationsResult{{Version: 0, IsApplied: true}},
		migrations...,
	)
}
