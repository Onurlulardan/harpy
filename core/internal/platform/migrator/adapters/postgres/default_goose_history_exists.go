package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/default_goose_history_exists.sql
var defaultGooseHistoryExistsSQL string

func DefaultGooseHistoryExists(
	ctx context.Context,
	database goosedb.DBTxConn,
) (bool, error) {
	var exists bool
	if err := database.QueryRowContext(ctx, defaultGooseHistoryExistsSQL).
		Scan(&exists); err != nil {
		return false, fmt.Errorf("inspect default Goose history: %w", err)
	}
	return exists, nil
}
