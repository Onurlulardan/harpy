package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *Store) GetMigration(
	ctx context.Context,
	db goosedb.DBTxConn,
	version int64,
) (*goosedb.GetMigrationResult, error) {
	var result goosedb.GetMigrationResult
	if err := db.QueryRowContext(ctx, getMigrationSQL, version).Scan(
		&result.Timestamp, &result.IsApplied,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %d", goosedb.ErrVersionNotFound, version)
		}
		return nil, fmt.Errorf("get migration history version %d: %w", version, err)
	}
	return &result, nil
}
