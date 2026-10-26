package postgres

import (
	"context"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorVerifyRelationsV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav2.IdentifierMapping,
	expected []schemav2.RelationArguments,
) error {
	for _, relation := range expected {
		if err := schemaExecutorVerifyRelationV2(
			ctx, querier, schemaName, mapping, relation,
		); err != nil {
			return err
		}
	}
	return nil
}

func schemaExecutorVerifyRelationV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav2.IdentifierMapping,
	expected schemav2.RelationArguments,
) error {
	table, err := schemaExecutorIdentifierNameV2(mapping.Relations, expected.RelationID)
	if err != nil {
		return err
	}
	if err = schemaExecutorRequireTable(ctx, querier, schemaName, table); err != nil {
		return err
	}
	columnCount, err := schemaExecutorTableCount(
		ctx, querier, schemaExecutorReadTableColumnCountSQL, schemaName, table,
	)
	if err != nil {
		return err
	}
	if columnCount != 4 {
		return schemaExecutorBlocked("schema_relation_column_mismatch", nil)
	}
	columns := map[string]schemaExecutorColumnContract{
		"source_id": {typeName: "uuid", notNull: true, defaultCollation: true},
		"target_id": {typeName: "uuid", notNull: true, defaultCollation: true},
		"edge_slot": {typeName: "smallint", notNull: true, defaultCollation: true},
		"created_at": {typeName: "timestamp with time zone", notNull: true,
			defaultExpression: "statement_timestamp()", defaultCollation: true},
	}
	for name, wanted := range columns {
		actual, exists, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, table, name,
		)
		if err != nil {
			return err
		}
		if !exists || actual != wanted {
			return schemaExecutorBlocked("schema_relation_column_mismatch", nil)
		}
	}
	constraintCount, err := schemaExecutorTableCount(
		ctx, querier, schemaExecutorReadTableConstraintCountSQL, schemaName, table,
	)
	if err != nil {
		return err
	}
	if constraintCount != 9 {
		return schemaExecutorBlocked("schema_relation_constraint_mismatch", nil)
	}
	indexCount, err := schemaExecutorTableCount(
		ctx, querier, schemaExecutorReadTableIndexCountSQL, schemaName, table,
	)
	if err != nil {
		return err
	}
	if indexCount != 2 {
		return schemaExecutorBlocked("schema_relation_index_count_mismatch", nil)
	}
	return schemaExecutorVerifyRelationKeysV2(
		ctx, querier, schemaName, table, mapping, expected,
	)
}
