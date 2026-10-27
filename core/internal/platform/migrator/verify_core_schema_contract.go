package migrator

import (
	"context"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	goosedb "github.com/pressly/goose/v3/database"
)

func (runner *Runner) verifyCoreSchemaContract(
	ctx context.Context,
	database goosedb.DBTxConn,
	actual postgresstore.CoreSchemaContractRecord,
) error {
	expected := runner.coreSchemaContract(actual.VerifiedRunID)
	if !sameCoreSchemaContract(actual, expected) {
		return fmt.Errorf("core schema contract differs from verified release")
	}
	expectedGeneration, err := runner.currentReleaseGeneration(actual.VerifiedRunID)
	if err != nil {
		return err
	}
	actualGeneration, err := postgresstore.GetReleaseGeneration(
		ctx, database, actual.VerifiedRunID,
	)
	if err != nil {
		return err
	}
	if actualGeneration != expectedGeneration {
		return fmt.Errorf("core schema contract release generation differs")
	}
	run, err := postgresstore.GetMigrationRun(ctx, database, actual.VerifiedRunID)
	if err != nil {
		return err
	}
	if err := runner.validateRunRecord(run); err != nil {
		return err
	}
	if run.State != "succeeded" {
		return fmt.Errorf("core schema contract references a non-succeeded run")
	}
	return nil
}

func sameCoreSchemaContract(
	actual, expected postgresstore.CoreSchemaContractRecord,
) bool {
	return actual.DatabaseGeneration == expected.DatabaseGeneration &&
		actual.MinimumRuntimeGeneration == expected.MinimumRuntimeGeneration &&
		actual.MigrationSHA256 == expected.MigrationSHA256 &&
		actual.ExtensionSHA256 == expected.ExtensionSHA256 &&
		actual.Status == expected.Status &&
		actual.VerifiedRunID == expected.VerifiedRunID && actual.Revision > 0
}
