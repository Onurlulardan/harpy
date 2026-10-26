package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/complete_migration_run.sql
var completeMigrationRunSQL string

func CompleteMigrationRun(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID string,
) error {
	result, err := database.ExecContext(ctx, completeMigrationRunSQL, runID)
	if err != nil {
		return fmt.Errorf("complete migration run: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return fmt.Errorf("complete migration run changed %d rows", affected)
	}
	return nil
}
