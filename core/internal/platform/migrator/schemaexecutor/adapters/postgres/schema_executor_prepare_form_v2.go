package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorPrepareFormV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	mapping, err := schemaExecutorMappingV2(step)
	if err != nil {
		return false, err
	}
	exists, err := schemaExecutorOwnedTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	)
	if err != nil {
		return false, err
	}
	switch step.Operation() {
	case schemav1.CreateFormTable:
		if exists {
			return false, schemaExecutorBlocked("schema_object_already_exists", nil)
		}
		return true, nil
	case schemav1.DropFormTable:
		return exists, nil
	}
	if !exists {
		return false, schemaExecutorBlocked("schema_form_table_missing", nil)
	}
	return schemaExecutorPrepareFieldV2(ctx, querier, schemaName, mapping, step)
}

func schemaExecutorPrepareFieldV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav2.IdentifierMapping,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	if step.Operation() == schemav1.OperationID(schemav2.SetFieldNotNull) {
		arguments, ok := step.ArgumentValue().(*schemav2.FieldArguments)
		if !ok {
			return false, fmt.Errorf("V2 not-null arguments are invalid")
		}
		name, err := schemaExecutorFieldNameV2(mapping, schemav2.FieldReference{
			FieldID: arguments.FieldID, Generation: arguments.Generation,
			Variant: arguments.Variant,
		})
		if err != nil {
			return false, err
		}
		actual, exists, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, mapping.Form.PhysicalName, name,
		)
		if err != nil || !exists {
			return false, schemaExecutorBlocked("schema_column_missing", err)
		}
		return !actual.notNull, nil
	}
	if step.Operation() == schemav1.DropFieldColumn {
		arguments, ok := step.ArgumentValue().(*schemav2.FieldArguments)
		if !ok {
			return false, fmt.Errorf("V2 field drop arguments are invalid")
		}
		name, err := schemaExecutorFieldNameV2(mapping, schemav2.FieldReference{
			FieldID: arguments.FieldID, Generation: arguments.Generation,
			Variant: arguments.Variant,
		})
		if err != nil {
			return false, err
		}
		_, exists, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, mapping.Form.PhysicalName, name,
		)
		return exists, err
	}
	arguments, ok := step.ArgumentValue().(*schemav2.ColumnArguments)
	if !ok || step.Operation() != schemav1.AddFieldColumn {
		return false, fmt.Errorf("V2 field add arguments are invalid")
	}
	name, err := schemaExecutorFieldNameV2(mapping, schemav2.FieldReference{
		FieldID: arguments.FieldID, Generation: arguments.Generation,
		Variant: arguments.Variant,
	})
	if err != nil {
		return false, err
	}
	actual, exists, err := schemaExecutorReadColumn(
		ctx, querier, schemaName, mapping.Form.PhysicalName, name,
	)
	if err != nil || !exists {
		return !exists, err
	}
	typeName, err := ddl.ColumnTypeV2(*arguments)
	if err != nil {
		return false, err
	}
	wanted := schemaExecutorColumnContract{
		typeName: typeName, notNull: !arguments.Nullable, defaultCollation: true,
	}
	if actual == wanted {
		return false, nil
	}
	return false, schemaExecutorBlocked("schema_column_contract_mismatch", nil)
}
