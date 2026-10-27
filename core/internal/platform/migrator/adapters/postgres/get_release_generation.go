package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/get_release_generation.sql
var getReleaseGenerationSQL string

func GetReleaseGeneration(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID string,
) (ReleaseGenerationRecord, error) {
	var record ReleaseGenerationRecord
	err := database.QueryRowContext(ctx, getReleaseGenerationSQL, runID).Scan(
		&record.MigrationRunID, &record.ReleaseID, &record.EnvironmentID,
		&record.DatabaseOrigin, &record.DatabaseGeneration,
		&record.MinimumRuntimeGeneration, &record.RuntimeGeneration,
		&record.MinimumDatabaseGeneration, &record.MigrationSHA256,
		&record.ExtensionSHA256, &record.PostgresImageDigest,
		&record.BootstrapReceiptRunID, &record.ExtensionReceiptRunID,
		&record.CleanupReceiptRunID,
	)
	if err != nil {
		return ReleaseGenerationRecord{}, fmt.Errorf("get release generation: %w", err)
	}
	return record, nil
}
