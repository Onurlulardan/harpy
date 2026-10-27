package postgres

import (
	"context"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *Store) ListMigrations(
	ctx context.Context,
	db goosedb.DBTxConn,
) ([]*goosedb.ListMigrationsResult, error) {
	rows, err := db.QueryContext(ctx, listMigrationsSQL)
	if err != nil {
		return nil, fmt.Errorf("list migration history: %w", err)
	}
	defer rows.Close()
	var migrations []*goosedb.ListMigrationsResult
	for rows.Next() {
		var migration goosedb.ListMigrationsResult
		if err := rows.Scan(&migration.Version, &migration.IsApplied); err != nil {
			return nil, fmt.Errorf("scan migration history: %w", err)
		}
		migrations = append(migrations, &migration)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate migration history: %w", err)
	}
	return migrations, nil
}
