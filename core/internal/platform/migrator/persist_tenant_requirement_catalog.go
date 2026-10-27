package migrator

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	goosedb "github.com/pressly/goose/v3/database"
)

func (runner *Runner) persistTenantRequirementCatalog(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID string,
) error {
	if len(runner.tenantSelected) == 0 {
		return nil
	}
	targets, err := postgresstore.ListTenantMigrationTargets(ctx, database)
	if err != nil {
		return err
	}
	expected, err := runner.buildTenantRequirementCatalog(runID, targets)
	if err != nil {
		return err
	}
	if err := postgresstore.CreateTenantRequirementCatalog(
		ctx, database, expected,
	); err != nil {
		return err
	}
	return runner.verifyTenantRequirementCatalogRecord(ctx, database, expected)
}

func (runner *Runner) verifyTenantRequirementCatalog(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID string,
) error {
	if len(runner.tenantSelected) == 0 {
		return nil
	}
	targets, err := postgresstore.ListTenantMigrationTargets(ctx, database)
	if err != nil {
		return err
	}
	expected, err := runner.buildTenantRequirementCatalog(runID, targets)
	if err != nil {
		return err
	}
	return runner.verifyTenantRequirementCatalogRecord(ctx, database, expected)
}

func (runner *Runner) verifyTenantRequirementCatalogRecord(
	ctx context.Context,
	database goosedb.DBTxConn,
	expected postgresstore.TenantRequirementCatalog,
) error {
	actual, err := postgresstore.GetTenantRequirementCatalog(
		ctx, database, expected.MigrationRunID,
	)
	if err != nil {
		return err
	}
	if actual.ContractVersion != expected.ContractVersion ||
		actual.MigrationRunID != expected.MigrationRunID ||
		actual.ReleaseID != expected.ReleaseID ||
		actual.EnvironmentID != expected.EnvironmentID ||
		actual.DatabaseOrigin != expected.DatabaseOrigin ||
		actual.MigrationHash != expected.MigrationHash ||
		actual.CatalogHash != expected.CatalogHash ||
		actual.RequirementCount != expected.RequirementCount ||
		actual.TargetCount != expected.TargetCount ||
		!bytes.Equal(actual.Canonical, expected.Canonical) ||
		!slices.Equal(actual.Requirements, expected.Requirements) ||
		!slices.Equal(actual.Targets, expected.Targets) {
		return fmt.Errorf("stored tenant requirement catalog differs")
	}
	return nil
}
