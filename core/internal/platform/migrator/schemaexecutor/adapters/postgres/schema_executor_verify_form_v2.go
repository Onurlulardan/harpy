package postgres

import (
	"context"
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorVerifyFormV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav2.IdentifierMapping,
	arguments schemav2.VerifyArguments,
) error {
	table := mapping.Form.PhysicalName
	if err := schemaExecutorRequireTable(ctx, querier, schemaName, table); err != nil {
		return err
	}
	actualColumns, err := schemaExecutorTableCount(
		ctx, querier, schemaExecutorReadTableColumnCountSQL, schemaName, table,
	)
	if err != nil {
		return err
	}
	expectedFields := 0
	if slices.Contains(arguments.Checks, schemav2.VerifyFields) {
		expectedFields = len(arguments.Fields)
	}
	if actualColumns != int64(schemav2.FormSystemColumnCount+expectedFields) {
		return schemaExecutorBlocked("schema_form_column_count_mismatch", nil)
	}
	columns := []struct {
		name     string
		contract schemaExecutorColumnContract
	}{
		{"id", schemaExecutorColumnContract{typeName: "uuid", notNull: true, defaultCollation: true}},
		{"created_at", schemaExecutorColumnContract{typeName: "timestamp with time zone", notNull: true,
			defaultExpression: "statement_timestamp()", defaultCollation: true}},
		{"created_by", schemaExecutorColumnContract{typeName: "uuid", notNull: true, defaultCollation: true}},
		{"updated_at", schemaExecutorColumnContract{typeName: "timestamp with time zone", notNull: true,
			defaultExpression: "statement_timestamp()", defaultCollation: true}},
		{"updated_by", schemaExecutorColumnContract{typeName: "uuid", notNull: true, defaultCollation: true}},
		{"row_version", schemaExecutorColumnContract{typeName: "bigint", notNull: true,
			defaultExpression: "1", defaultCollation: true}},
		{"deleted_at", schemaExecutorColumnContract{typeName: "timestamp with time zone", defaultCollation: true}},
		{"deleted_by", schemaExecutorColumnContract{typeName: "uuid", defaultCollation: true}},
	}
	for _, column := range columns {
		actual, exists, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, table, column.name,
		)
		if err != nil {
			return err
		}
		if !exists || actual != column.contract {
			return schemaExecutorBlocked("schema_form_contract_mismatch", nil)
		}
	}
	expectedConstraints := int64(-1)
	if slices.Contains(arguments.Checks, schemav2.VerifyConstraints) &&
		slices.Contains(arguments.Checks, schemav2.VerifyFields) {
		expectedConstraints = 14 + int64(len(arguments.Constraints))
		for _, field := range arguments.Fields {
			if !field.Nullable {
				expectedConstraints++
			}
		}
	}
	if err = schemaExecutorVerifyFormConstraintsV2(
		ctx, querier, schemaName, table, expectedConstraints,
	); err != nil {
		return err
	}
	return schemaExecutorVerifyFormIndexCountV2(
		ctx, querier, schemaName, table, arguments,
	)
}

func schemaExecutorVerifyFormIndexCountV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	arguments schemav2.VerifyArguments,
) error {
	if !slices.Contains(arguments.Checks, schemav2.VerifyIndexes) ||
		!slices.Contains(arguments.Checks, schemav2.VerifyConstraints) {
		return nil
	}
	expected := int64(1 + len(arguments.Indexes))
	for _, constraint := range arguments.Constraints {
		if constraint.Kind == schemav2.ConstraintUnique {
			expected++
		}
	}
	actual, err := schemaExecutorTableCount(
		ctx, querier, schemaExecutorReadTableIndexCountSQL, schemaName, tableName,
	)
	if err != nil {
		return err
	}
	if actual != expected {
		return schemaExecutorBlocked("schema_form_index_count_mismatch", nil)
	}
	return nil
}
