package migrationmanifest

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func newLogicalMigration(
	group SourceGroup,
	identity documentIdentity,
	transactionMode string,
) *logicalMigration {
	return &logicalMigration{
		logicalName: identity.logicalName, transactionMode: transactionMode,
		entry: postgresv1.MigrationEntry{
			Owner: group.Owner, Scope: group.Scope, Version: identity.version,
			TransactionMode: transactionMode, Phase: group.Phase,
			TargetDatabaseGeneration:     group.TargetDatabaseGeneration,
			NextMinimumRuntimeGeneration: group.NextMinimumRuntimeGeneration,
		},
	}
}
