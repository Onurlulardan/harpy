package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/release_generation_table_exists.sql
var releaseGenerationTableExistsSQL string

func ReleaseGenerationTableExists(
	ctx context.Context,
	database goosedb.DBTxConn,
) (bool, error) {
	var exists bool
	if err := database.QueryRowContext(
		ctx, releaseGenerationTableExistsSQL,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check release generation table: %w", err)
	}
	return exists, nil
}
