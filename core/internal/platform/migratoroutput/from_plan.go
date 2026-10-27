package migratoroutput

import (
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
)

func FromPlan(value platformmigrator.PlanResult) (Result, error) {
	migrations, err := canonicalPlanMigrations(value.Pending)
	if err != nil {
		return Result{}, err
	}
	pending := make([]runtimev1.MigrationPlanEntry, 0, len(migrations))
	for _, migration := range migrations {
		pending = append(pending, runtimev1.MigrationPlanEntry{
			Owner: migration.Owner, Scope: migration.Scope, Version: migration.Version,
			TransactionMode: migration.TransactionMode, Phase: migration.Phase,
			TargetDatabaseGeneration:     migration.TargetDatabaseGeneration,
			NextMinimumRuntimeGeneration: migration.NextMinimumRuntimeGeneration,
			Origin:                       string(migration.Origin), Path: migration.Path, SHA256: migration.SHA256,
		})
	}
	return newResult(runtimev1.MigratorCommandResult{
		ContractVersion: runtimev1.MigratorCommandResultVersion,
		Success:         true, Command: "plan",
		Plan: &runtimev1.MigrationPlanResult{
			CurrentVersion: value.CurrentVersion, Pending: pending,
		},
	})
}
