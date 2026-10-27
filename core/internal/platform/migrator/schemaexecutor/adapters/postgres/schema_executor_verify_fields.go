package postgres

import (
	"context"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
)

func schemaExecutorVerifyFields(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav1.IdentifierMapping,
	expected []schemav1.VerifyFieldContract,
) error {
	for _, field := range expected {
		name, mapped := schemaExecutorFieldGenerationName(
			mapping, field.FieldID, field.Generation,
		)
		columnType, err := ddl.ColumnType(schemav1.ColumnArguments{
			StorageType:      field.StorageType,
			DecimalPrecision: field.DecimalPrecision, DecimalScale: field.DecimalScale,
		})
		if err != nil || !mapped {
			return schemaExecutorBlocked("schema_field_contract_invalid", err)
		}
		actual, exists, err := schemaExecutorReadColumn(
			ctx, querier, schemaName, mapping.Form.PhysicalName, name,
		)
		if err != nil {
			return err
		}
		wanted := schemaExecutorColumnContract{
			typeName: columnType, notNull: !field.Nullable, defaultCollation: true,
		}
		if !exists || actual != wanted {
			return schemaExecutorBlocked("schema_field_contract_mismatch", nil)
		}
	}
	return nil
}

func schemaExecutorVerifyRelations(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav1.IdentifierMapping,
	expected []schemav1.RelationArguments,
) error {
	for _, contract := range expected {
		if err := schemaExecutorVerifyRelation(
			ctx, querier, schemaName, mapping, contract,
		); err != nil {
			return err
		}
	}
	return nil
}
