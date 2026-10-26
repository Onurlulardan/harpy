package postgres

import (
	"context"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *Store) TableExists(
	ctx context.Context,
	db goosedb.DBTxConn,
) (bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, tableExistsSQL).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration history table: %w", err)
	}
	return exists, nil
}
