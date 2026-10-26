package postgres

import (
	"context"
	"slices"
)

func schemaExecutorVerifyFormConstraints(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	customConstraintCount int,
) error {
	if customConstraintCount >= 0 {
		count, err := schemaExecutorTableCount(
			ctx, querier, schemaExecutorReadTableConstraintCountSQL, schemaName, tableName,
		)
		if err != nil {
			return err
		}
		if count != int64(2+customConstraintCount) {
			return schemaExecutorBlocked("schema_form_constraint_count_mismatch", nil)
		}
	}
	primary, found, err := schemaExecutorReadKey(
		ctx, querier, schemaName, tableName, tableName+"_pkey",
	)
	if err != nil {
		return err
	}
	if !found || primary.owner != "harpy_owner" ||
		primary.kind != "p" || !primary.validated ||
		primary.deferrable || primary.deferred ||
		!slices.Equal(primary.sourceFields, []string{"id"}) {
		return schemaExecutorBlocked("schema_form_primary_key_mismatch", nil)
	}
	rowVersion, found, err := schemaExecutorReadConstraint(
		ctx, querier, schemaName, tableName, tableName+"_row_version_check",
	)
	if err != nil {
		return err
	}
	if !found || rowVersion.owner != "harpy_owner" ||
		rowVersion.kind != "c" || !rowVersion.validated ||
		schemaExecutorNormalizeDefinition(rowVersion.definition) != "checkrow_version>0" {
		return schemaExecutorBlocked("schema_form_row_version_mismatch", nil)
	}
	return nil
}
