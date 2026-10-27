package postgres

import (
	"context"
	"slices"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorVerifyStorage(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	trigramSchema string,
	step schemaexecutor.ValidatedStep,
) error {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return schemaExecutorVerifyStorageV2(
			ctx, querier, schemaName, trigramSchema, step,
		)
	}
	arguments := step.ArgumentValue().(*schemav1.VerifyArguments)
	for _, check := range arguments.Checks {
		var err error
		switch check {
		case schemav1.VerifyFormTable:
			fieldCount, constraintCount := -1, -1
			if slices.Contains(arguments.Checks, schemav1.VerifyFields) {
				fieldCount = len(arguments.Fields)
			}
			if slices.Contains(arguments.Checks, schemav1.VerifyConstraints) {
				constraintCount = len(arguments.Constraints)
			}
			err = schemaExecutorVerifyForm(
				ctx, querier, schemaName, step.Mapping(),
				fieldCount, constraintCount,
			)
		case schemav1.VerifyFields:
			err = schemaExecutorVerifyFields(
				ctx, querier, schemaName, step.Mapping(), arguments.Fields,
			)
		case schemav1.VerifyRelations:
			err = schemaExecutorVerifyRelations(
				ctx, querier, schemaName, step.Mapping(), arguments.Relations,
			)
		case schemav1.VerifyIndexes:
			err = schemaExecutorVerifyIndexes(
				ctx, querier, schemaName, step.Mapping(), arguments.Indexes,
			)
		case schemav1.VerifyConstraints:
			err = schemaExecutorVerifyConstraints(
				ctx, querier, schemaName, step.Mapping(), arguments.Constraints,
			)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func schemaExecutorVerifyForm(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav1.IdentifierMapping,
	fieldCount int,
	constraintCount int,
) error {
	if err := schemaExecutorRequireTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	); err != nil {
		return err
	}
	if fieldCount >= 0 {
		actualColumns, err := schemaExecutorTableCount(
			ctx, querier, schemaExecutorReadTableColumnCountSQL,
			schemaName, mapping.Form.PhysicalName,
		)
		if err != nil {
			return err
		}
		if actualColumns != int64(5+fieldCount) {
			return schemaExecutorBlocked("schema_form_column_count_mismatch", nil)
		}
	}
	expected := []schemaExecutorColumnContract{
		{typeName: "uuid", notNull: true, defaultCollation: true},
		{typeName: "timestamp with time zone", notNull: true,
			defaultExpression: "statement_timestamp()", defaultCollation: true},
		{typeName: "timestamp with time zone", notNull: true,
			defaultExpression: "statement_timestamp()", defaultCollation: true},
		{typeName: "bigint", notNull: true,
			defaultExpression: "1", defaultCollation: true},
		{typeName: "timestamp with time zone", notNull: false, defaultCollation: true},
	}
	for index, name := range []string{"id", "created_at", "updated_at", "row_version", "deleted_at"} {
		actual, exists, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, mapping.Form.PhysicalName, name,
		)
		if err != nil {
			return err
		}
		if !exists || actual != expected[index] {
			return schemaExecutorBlocked("schema_form_contract_mismatch", nil)
		}
	}
	return schemaExecutorVerifyFormConstraints(
		ctx, querier, schemaName, mapping.Form.PhysicalName, constraintCount,
	)
}
