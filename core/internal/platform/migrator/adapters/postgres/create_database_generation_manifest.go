package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/create_database_generation_manifest.sql
var createDatabaseGenerationManifestSQL string

func CreateDatabaseGenerationManifest(
	ctx context.Context,
	database goosedb.DBTxConn,
	record DatabaseGenerationManifestRecord,
) error {
	result, err := database.ExecContext(
		ctx, createDatabaseGenerationManifestSQL,
		record.EnvironmentID, record.DatabaseGeneration,
		record.MigrationSHA256, record.ExtensionSHA256,
	)
	if err != nil {
		return fmt.Errorf("create database generation manifest: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected > 1 {
		return fmt.Errorf("create database generation manifest returned an invalid row count")
	}
	return nil
}
