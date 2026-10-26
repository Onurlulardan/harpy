package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/get_migration_run.sql
var getMigrationRunSQL string

func GetMigrationRun(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID string,
) (MigrationRunRecord, error) {
	var record MigrationRunRecord
	if err := database.QueryRowContext(ctx, getMigrationRunSQL, runID).Scan(
		&record.RunID, &record.ReleaseID, &record.EnvironmentID,
		&record.DatabaseOrigin, &record.Command, &record.State,
		&record.MigrationManifestSHA256, &record.ExtensionManifestSHA256,
		&record.RolesManifestSHA256, &record.PrivilegesManifestSHA256,
		&record.OwnershipManifestSHA256, &record.FailureCode,
	); err != nil {
		return MigrationRunRecord{}, fmt.Errorf("get migration run: %w", err)
	}
	return record, nil
}
