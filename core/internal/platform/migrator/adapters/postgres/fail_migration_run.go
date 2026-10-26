package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/fail_migration_run.sql
var failMigrationRunSQL string

func FailMigrationRun(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID, failureCode string,
) error {
	result, err := database.ExecContext(ctx, failMigrationRunSQL, runID, failureCode)
	if err != nil {
		return fmt.Errorf("fail migration run: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return fmt.Errorf("fail migration run changed %d rows", affected)
	}
	return nil
}
