package migratoroutput

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

func NewBaselineUp(runID string, applied []int64, version int64) (Result, error) {
	if applied == nil {
		applied = []int64{}
	}
	return newResult(runtimev1.MigratorCommandResult{
		ContractVersion: runtimev1.MigratorCommandResultVersion,
		Success:         true, Command: "up",
		Up: &runtimev1.MigrationUpResult{
			RunID: runID, AppliedVersions: applied, DatabaseVersion: version,
		},
	})
}
