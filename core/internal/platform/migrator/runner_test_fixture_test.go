package migrator

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

const runnerTestRunID = "018f0000-0000-7000-8000-000000000001"

func runnerTestSelection() []postgresv1.SelectedMigration {
	return []postgresv1.SelectedMigration{
		{
			Owner: "platform_migrator", Scope: "platform",
			Version: runLifecycleReadyVersion, TransactionMode: "transactional",
			Phase: "foundation", TargetDatabaseGeneration: 1,
			NextMinimumRuntimeGeneration: 1,
			Origin:                       postgresv1.DatabaseOriginFresh,
			Path:                         "migrations/20260713013200_guard_migration_run_transitions.fresh.sql",
			SHA256:                       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			Owner: "platform_migrator", Scope: "platform",
			Version: 20260713013300, TransactionMode: "transactional",
			Phase: "foundation", TargetDatabaseGeneration: 1,
			NextMinimumRuntimeGeneration: 1,
			Origin:                       postgresv1.DatabaseOriginFresh,
			Path:                         "migrations/20260713013300_create_probe.fresh.sql",
			SHA256:                       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		},
	}
}

func runnerTestHistory(migration postgresv1.SelectedMigration) historyRow {
	return historyRow{
		version: migration.Version, recordKind: "migration", isApplied: true,
		owner: migration.Owner, scope: migration.Scope, origin: string(migration.Origin),
		selectedPath: migration.Path, selectedSHA256: migration.SHA256,
		runID:                        runnerTestRunID,
		targetDatabaseGeneration:     migration.TargetDatabaseGeneration,
		nextMinimumRuntimeGeneration: migration.NextMinimumRuntimeGeneration,
	}
}
