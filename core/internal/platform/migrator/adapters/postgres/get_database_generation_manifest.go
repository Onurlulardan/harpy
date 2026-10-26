package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/get_database_generation_manifest.sql
var getDatabaseGenerationManifestSQL string

func GetDatabaseGenerationManifest(
	ctx context.Context,
	database goosedb.DBTxConn,
	environmentID string,
	databaseGeneration int64,
) (DatabaseGenerationManifestRecord, error) {
	var record DatabaseGenerationManifestRecord
	err := database.QueryRowContext(
		ctx, getDatabaseGenerationManifestSQL, environmentID, databaseGeneration,
	).Scan(
		&record.EnvironmentID, &record.DatabaseGeneration,
		&record.MigrationSHA256, &record.ExtensionSHA256,
	)
	if err != nil {
		return DatabaseGenerationManifestRecord{},
			fmt.Errorf("get database generation manifest: %w", err)
	}
	return record, nil
}
