package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/start_migration_run.sql
var startMigrationRunSQL string

func StartMigrationRun(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID string,
) error {
	result, err := database.ExecContext(ctx, startMigrationRunSQL, runID)
	if err != nil {
		return fmt.Errorf("start migration run: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return fmt.Errorf("start migration run changed %d rows", affected)
	}
	return nil
}
