package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorPrepareForm(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return schemaExecutorPrepareFormV2(ctx, querier, schemaName, step)
	}
	mapping := step.Mapping()
	table := mapping.Form.PhysicalName
	exists, err := schemaExecutorOwnedTable(ctx, querier, schemaName, table)
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
	return schemaExecutorPrepareField(ctx, querier, schemaName, step)
}

func schemaExecutorPrepareField(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	mapping := step.Mapping()
	switch step.Operation() {
	case schemav1.AddFieldColumn:
		arguments := step.ArgumentValue().(*schemav1.ColumnArguments)
		column, err := schemaExecutorFieldName(mapping, arguments.FieldID)
		if err != nil {
			return false, err
		}
		contract, exists, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, mapping.Form.PhysicalName, column,
		)
		if err != nil || !exists {
			return !exists, err
		}
		typeName, err := ddl.ColumnType(*arguments)
		if err != nil {
			return false, err
		}
		if contract.typeName == typeName && contract.notNull == !arguments.Nullable &&
			contract.defaultExpression == "" && contract.defaultCollation {
			return false, nil
		}
		return false, schemaExecutorBlocked("schema_column_contract_mismatch", nil)
	case schemav1.DropFieldColumn:
		arguments := step.ArgumentValue().(*schemav1.FieldArguments)
		column, err := schemaExecutorFieldName(mapping, arguments.FieldID)
		if err != nil {
			return false, err
		}
		_, exists, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, mapping.Form.PhysicalName, column,
		)
		return exists, err
	default:
		return false, fmt.Errorf("form operation precondition is not registered")
	}
}
