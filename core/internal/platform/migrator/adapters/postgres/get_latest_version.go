package postgres

import (
	"context"
	"database/sql"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *Store) GetLatestVersion(
	ctx context.Context,
	db goosedb.DBTxConn,
) (int64, error) {
	var version sql.NullInt64
	if err := db.QueryRowContext(ctx, getLatestVersionSQL).Scan(&version); err != nil {
		return -1, fmt.Errorf("get latest migration history version: %w", err)
	}
	if !version.Valid {
		return -1, fmt.Errorf("latest %w", goosedb.ErrVersionNotFound)
	}
	return version.Int64, nil
}
