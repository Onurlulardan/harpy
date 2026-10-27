package migrator

import (
	"context"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	goosedb "github.com/pressly/goose/v3/database"
)

func (runner *Runner) persistDatabaseGenerationManifest(
	ctx context.Context,
	database goosedb.DBTxConn,
) error {
	expected := postgresstore.DatabaseGenerationManifestRecord{
		EnvironmentID:      runner.identity.environmentID,
		DatabaseGeneration: runner.identity.databaseGeneration,
		MigrationSHA256:    runner.identity.manifestSHA256,
		ExtensionSHA256:    runner.identity.extensionSHA256,
	}
	if err := postgresstore.CreateDatabaseGenerationManifest(
		ctx, database, expected,
	); err != nil {
		return err
	}
	actual, err := postgresstore.GetDatabaseGenerationManifest(
		ctx, database, expected.EnvironmentID, expected.DatabaseGeneration,
	)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("database generation manifest fork")
	}
	return nil
}
