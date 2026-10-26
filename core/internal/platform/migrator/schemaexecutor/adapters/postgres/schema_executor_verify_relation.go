package postgres

import (
	"context"
	"slices"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func schemaExecutorVerifyRelation(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav1.IdentifierMapping,
	expected schemav1.RelationArguments,
) error {
	tableName, err := schemaExecutorIdentifierName(mapping.Relations, expected.RelationID)
	if err != nil {
		return err
	}
	if err := schemaExecutorRequireTable(ctx, querier, schemaName, tableName); err != nil {
		return err
	}
	columnCount, err := schemaExecutorTableCount(
		ctx, querier, schemaExecutorReadTableColumnCountSQL, schemaName, tableName,
	)
	if err != nil {
		return err
	}
	if columnCount != 3 {
		return schemaExecutorBlocked("schema_relation_column_mismatch", nil)
	}
	columns := map[string]schemaExecutorColumnContract{
		"source_id": {typeName: "uuid", notNull: true, defaultCollation: true},
		"target_id": {typeName: "uuid", notNull: true, defaultCollation: true},
		"created_at": {typeName: "timestamp with time zone", notNull: true,
			defaultExpression: "statement_timestamp()", defaultCollation: true},
	}
	for name, wanted := range columns {
		actual, found, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, tableName, name,
		)
		if err != nil {
			return err
		}
		if !found || actual != wanted {
			return schemaExecutorBlocked("schema_relation_column_mismatch", nil)
		}
	}
	constraintCount, err := schemaExecutorTableCount(
		ctx, querier, schemaExecutorReadTableConstraintCountSQL, schemaName, tableName,
	)
	expectedCount := int64(3)
	if expected.CardinalityConstraintID != nil {
		expectedCount++
	}
	if err != nil {
		return err
	}
	if constraintCount != expectedCount {
		return schemaExecutorBlocked("schema_relation_constraint_mismatch", nil)
	}
	return schemaExecutorVerifyRelationKeys(
		ctx, querier, schemaName, tableName, mapping, expected,
	)
}

func schemaExecutorVerifyRelationKeys(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	mapping schemav1.IdentifierMapping,
	expected schemav1.RelationArguments,
) error {
	primary, found, err := schemaExecutorReadKey(
		ctx, querier, schemaName, tableName, tableName+"_pkey",
	)
	if err != nil {
		return err
	}
	if !found || primary.owner != "harpy_owner" ||
		primary.kind != "p" || !primary.validated ||
		primary.deferrable || primary.deferred ||
		!slices.Equal(primary.sourceFields, []string{"source_id", "target_id"}) {
		return schemaExecutorBlocked("schema_relation_primary_key_mismatch", nil)
	}
	if err := schemaExecutorVerifyForeignKeys(
		ctx, querier, schemaName, tableName, mapping, expected,
	); err != nil {
		return err
	}
	return schemaExecutorVerifyCardinalityKey(
		ctx, querier, schemaName, tableName, mapping, expected,
	)
}
