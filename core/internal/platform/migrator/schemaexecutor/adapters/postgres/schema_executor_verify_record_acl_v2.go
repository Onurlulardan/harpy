package postgres

import (
	"context"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorVerifyRecordACLsV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav2.IdentifierMapping,
) error {
	tables := make([]string, 0, 1+len(mapping.Relations))
	tables = append(tables, mapping.Form.PhysicalName)
	for _, relation := range mapping.Relations {
		tables = append(tables, relation.PhysicalName)
	}
	for _, table := range tables {
		if err := schemaExecutorRequireTable(ctx, querier, schemaName, table); err != nil {
			return err
		}
		var tableACLValid, schemaACLValid bool
		err := querier.QueryRow(
			ctx, schemaExecutorReadRecordACLSQL, schemaName, table,
		).Scan(&tableACLValid, &schemaACLValid)
		if err != nil {
			return err
		}
		if !tableACLValid || !schemaACLValid {
			return schemaExecutorBlocked("schema_record_acl_mismatch", nil)
		}
	}
	return nil
}
