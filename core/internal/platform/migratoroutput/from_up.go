package migratoroutput

import (
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
)

func FromUp(value platformmigrator.UpResult) (Result, error) {
	applied := canonicalAppliedVersions(value.AppliedVersions)
	return newResult(runtimev1.MigratorCommandResult{
		ContractVersion: runtimev1.MigratorCommandResultVersion,
		Success:         true, Command: "up",
		Up: &runtimev1.MigrationUpResult{
			RunID: value.RunID, AppliedVersions: applied,
			DatabaseVersion: value.DatabaseVersion,
		},
	})
}
