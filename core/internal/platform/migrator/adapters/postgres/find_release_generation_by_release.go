package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/find_release_generation_by_release.sql
var findReleaseGenerationByReleaseSQL string

func FindReleaseGenerationByRelease(
	ctx context.Context,
	database goosedb.DBTxConn,
	environmentID string,
	releaseID string,
) (ReleaseGenerationRecord, error) {
	var record ReleaseGenerationRecord
	err := database.QueryRowContext(
		ctx, findReleaseGenerationByReleaseSQL, environmentID, releaseID,
	).Scan(
		&record.MigrationRunID, &record.ReleaseID, &record.EnvironmentID,
		&record.DatabaseOrigin, &record.DatabaseGeneration,
		&record.MinimumRuntimeGeneration, &record.RuntimeGeneration,
		&record.MinimumDatabaseGeneration, &record.MigrationSHA256,
		&record.ExtensionSHA256, &record.PostgresImageDigest,
		&record.BootstrapReceiptRunID, &record.ExtensionReceiptRunID,
		&record.CleanupReceiptRunID,
	)
	if err != nil {
		return ReleaseGenerationRecord{}, fmt.Errorf("find release generation: %w", err)
	}
	return record, nil
}
