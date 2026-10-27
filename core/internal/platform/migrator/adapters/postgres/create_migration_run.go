package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/create_migration_run.sql
var createMigrationRunSQL string

func CreateMigrationRun(
	ctx context.Context,
	database goosedb.DBTxConn,
	input CreateMigrationRunInput,
) error {
	result, err := database.ExecContext(
		ctx, createMigrationRunSQL,
		input.RunID, input.ReleaseID, input.EnvironmentID, input.DatabaseOrigin,
		input.Command, input.MigrationManifestSHA256, input.ExtensionManifestSHA256,
		input.RolesManifestSHA256, input.PrivilegesManifestSHA256,
		input.OwnershipManifestSHA256,
	)
	if err != nil {
		return fmt.Errorf("create migration run: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected > 1 {
		return fmt.Errorf("create migration run returned an invalid row count")
	}
	return nil
}
