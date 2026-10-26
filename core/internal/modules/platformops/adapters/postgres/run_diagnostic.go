package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/diagnostics"
)

func (store *Store) RunDiagnostic(
	ctx context.Context,
	definition diagnostics.Definition,
) ([]diagnostics.Field, error) {
	statement, names, err := diagnosticStatement(definition.ID)
	if err != nil {
		return nil, err
	}
	values := make([]string, len(names))
	scan := make([]any, len(values))
	for index := range values {
		scan[index] = &values[index]
	}
	err = store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		return database.QueryRow(ctx, statement).Scan(scan...)
	})
	if err != nil {
		return nil, fmt.Errorf("run registered database diagnostic: %w", err)
	}
	fields := make([]diagnostics.Field, len(names))
	for index := range names {
		fields[index] = diagnostics.Field{Name: names[index], Value: values[index]}
	}
	return fields, nil
}

func diagnosticStatement(
	id diagnostics.ID,
) (string, []string, error) {
	switch id {
	case diagnostics.DatabaseStats:
		return diagnosticDatabaseStatsSQL,
			[]string{"database", "database_bytes", "server_version_num", "read_only"}, nil
	case diagnostics.SchemaMetadata:
		return diagnosticSchemaMetadataSQL,
			[]string{"schema_count", "platform_count", "tenant_count", "last_schema"}, nil
	case diagnostics.RecoveryReadiness:
		return diagnosticRecoveryReadinessSQL,
			[]string{"storage_present", "signing_generation", "protected_sets", "read_only"}, nil
	default:
		return "", nil, diagnostics.ErrUnknownID
	}
}
