package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/create_release_generation.sql
var createReleaseGenerationSQL string

func CreateReleaseGeneration(
	ctx context.Context,
	database goosedb.DBTxConn,
	record ReleaseGenerationRecord,
) error {
	result, err := database.ExecContext(
		ctx, createReleaseGenerationSQL,
		record.MigrationRunID, record.ReleaseID, record.EnvironmentID,
		record.DatabaseOrigin, record.DatabaseGeneration,
		record.MinimumRuntimeGeneration, record.RuntimeGeneration,
		record.MinimumDatabaseGeneration, record.MigrationSHA256,
		record.ExtensionSHA256, record.PostgresImageDigest,
		record.BootstrapReceiptRunID, record.ExtensionReceiptRunID,
		record.CleanupReceiptRunID,
	)
	if err != nil {
		return fmt.Errorf("create release generation: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected > 1 {
		return fmt.Errorf("create release generation returned an invalid row count")
	}
	return nil
}
