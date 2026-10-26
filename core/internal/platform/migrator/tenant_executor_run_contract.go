package migrator

import (
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	tenantexec "github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
)

func (executor *TenantExecutor) runInput(
	claim tenantexec.Claim,
) postgresstore.CreateMigrationRunInput {
	release := executor.release
	return postgresstore.CreateMigrationRunInput{
		RunID: claim.MigrationRunID.String(), ReleaseID: release.ReleaseID,
		EnvironmentID:            claim.EnvironmentID.String(),
		DatabaseOrigin:           release.DatabaseOrigin,
		Command:                  "tenant_migration_executor",
		MigrationManifestSHA256:  release.MigrationManifestSHA256,
		ExtensionManifestSHA256:  release.ExtensionManifestSHA256,
		RolesManifestSHA256:      release.RolesManifestSHA256,
		PrivilegesManifestSHA256: release.PrivilegesManifestSHA256,
		OwnershipManifestSHA256:  release.OwnershipManifestSHA256,
	}
}

func (executor *TenantExecutor) validateRun(
	record postgresstore.MigrationRunRecord, claim tenantexec.Claim,
) error {
	expected := executor.runInput(claim)
	if record.RunID != expected.RunID || record.ReleaseID != expected.ReleaseID ||
		record.EnvironmentID != expected.EnvironmentID ||
		record.DatabaseOrigin != expected.DatabaseOrigin ||
		record.Command != expected.Command ||
		record.MigrationManifestSHA256 != expected.MigrationManifestSHA256 ||
		record.ExtensionManifestSHA256 != expected.ExtensionManifestSHA256 ||
		record.RolesManifestSHA256 != expected.RolesManifestSHA256 ||
		record.PrivilegesManifestSHA256 != expected.PrivilegesManifestSHA256 ||
		record.OwnershipManifestSHA256 != expected.OwnershipManifestSHA256 {
		return fmt.Errorf("tenant migration run differs from verified release")
	}
	if record.State != "planned" && record.State != "running" &&
		record.State != "succeeded" {
		return fmt.Errorf("tenant migration run is terminal")
	}
	return nil
}
