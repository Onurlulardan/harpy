package migratoroutput

import (
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
)

func FromVerify(value platformmigrator.VerifyResult) (Result, error) {
	return newResult(runtimev1.MigratorCommandResult{
		ContractVersion: runtimev1.MigratorCommandResultVersion,
		Success:         true, Command: "verify",
		Verify: &runtimev1.MigrationVerifyResult{
			DatabaseVersion:          value.DatabaseVersion,
			DatabaseGeneration:       value.DatabaseGeneration,
			MinimumRuntimeGeneration: value.MinimumRuntimeGeneration,
			VerifiedRunID:            value.VerifiedRunID,
		},
	})
}
