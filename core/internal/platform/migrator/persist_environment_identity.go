package migrator

import (
	"context"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	goosedb "github.com/pressly/goose/v3/database"
)

func (runner *Runner) persistEnvironmentIdentity(
	ctx context.Context,
	database goosedb.DBTxConn,
) error {
	expected := postgresstore.EnvironmentIdentityRecord{
		EnvironmentID:            runner.identity.environmentID,
		DatabaseOrigin:           string(runner.identity.databaseOrigin),
		DatabaseSystemID:         runner.identity.databaseSystemID,
		DatabaseOID:              runner.identity.databaseOID,
		BootstrapReleaseID:       runner.identity.releaseID,
		BootstrapAuthoritySHA256: runner.identity.authoritySHA256,
	}
	if err := postgresstore.CreateEnvironmentIdentity(ctx, database, expected); err != nil {
		return err
	}
	actual, err := postgresstore.GetEnvironmentIdentity(ctx, database)
	if err != nil {
		return err
	}
	if !sameEnvironmentIdentity(actual, expected) {
		return fmt.Errorf("stored environment identity differs from verified evidence")
	}
	return nil
}

func sameEnvironmentIdentity(
	actual, current postgresstore.EnvironmentIdentityRecord,
) bool {
	if actual.EnvironmentID != current.EnvironmentID ||
		actual.DatabaseOrigin != current.DatabaseOrigin ||
		actual.DatabaseSystemID != current.DatabaseSystemID ||
		actual.DatabaseOID != current.DatabaseOID {
		return false
	}
	if actual.BootstrapReleaseID == current.BootstrapReleaseID &&
		actual.BootstrapAuthoritySHA256 != current.BootstrapAuthoritySHA256 {
		return false
	}
	return actual.BootstrapReleaseID != "" &&
		actual.BootstrapAuthoritySHA256 != ""
}
