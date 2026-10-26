package postgres

import (
	"context"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
)

func schemaExecutorVerifyFieldsV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav2.IdentifierMapping,
	expected []schemav2.VerifyFieldContract,
) error {
	for _, field := range expected {
		reference := schemav2.FieldReference{
			FieldID: field.FieldID, Generation: field.Generation, Variant: field.Variant,
		}
		name, err := schemaExecutorFieldNameV2(mapping, reference)
		if err != nil {
			return schemaExecutorBlocked("schema_field_contract_invalid", err)
		}
		columnType, err := ddl.ColumnTypeV2(schemav2.ColumnArguments{
			StorageType: field.StorageType, DecimalPrecision: field.DecimalPrecision,
			DecimalScale: field.DecimalScale,
		})
		if err != nil {
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
