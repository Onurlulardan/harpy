package migratoroutput

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

func NewBaselinePlan(current int64, pending []runtimev1.MigrationPlanEntry) (Result, error) {
	if pending == nil {
		pending = []runtimev1.MigrationPlanEntry{}
	}
	return newResult(runtimev1.MigratorCommandResult{
		ContractVersion: runtimev1.MigratorCommandResultVersion,
		Success:         true, Command: "plan",
		Plan: &runtimev1.MigrationPlanResult{
			CurrentVersion: current, Pending: pending,
		},
	})
}

func NewBaselineVerify(runID string, version int64) (Result, error) {
	return newResult(runtimev1.MigratorCommandResult{
		ContractVersion: runtimev1.MigratorCommandResultVersion,
		Success:         true, Command: "verify",
		Verify: &runtimev1.MigrationVerifyResult{
			DatabaseVersion: version, DatabaseGeneration: 1,
			MinimumRuntimeGeneration: 1, VerifiedRunID: runID,
		},
	})
}
