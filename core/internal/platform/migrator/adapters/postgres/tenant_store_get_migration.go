package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *TenantStore) GetMigration(
	ctx context.Context,
	database goosedb.DBTxConn,
	version int64,
) (*goosedb.GetMigrationResult, error) {
	if version == 0 {
		return &goosedb.GetMigrationResult{
			Timestamp: time.Unix(0, 0).UTC(), IsApplied: true,
		}, nil
	}
	var result goosedb.GetMigrationResult
	err := database.QueryRowContext(
		ctx, getTenantStoreMigrationSQL, store.target.TenantID, version,
	).Scan(&result.Timestamp, &result.IsApplied)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %d", goosedb.ErrVersionNotFound, version)
	}
	if err != nil {
		return nil, fmt.Errorf("get tenant migration %d: %w", version, err)
	}
	return &result, nil
}
