package migrator

import (
	"context"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	goosedb "github.com/pressly/goose/v3/database"
)

func (runner *Runner) persistReleaseGeneration(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID string,
) error {
	expected, err := runner.currentReleaseGeneration(runID)
	if err != nil {
		return err
	}
	if err := postgresstore.CreateReleaseGeneration(ctx, database, expected); err != nil {
		return err
	}
	actual, err := postgresstore.GetReleaseGeneration(ctx, database, runID)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("stored release generation differs from verified release")
	}
	return nil
}
