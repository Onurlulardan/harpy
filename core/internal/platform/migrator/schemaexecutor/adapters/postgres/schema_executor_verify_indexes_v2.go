package postgres

import (
	"context"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorVerifyIndexesV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	trigramSchema string,
	mapping schemav2.IdentifierMapping,
	expected []schemav2.VerifyIndexContract,
) error {
	for _, index := range expected {
		name, err := schemaExecutorIndexNameV2(
			mapping.Indexes, index.IndexID, index.Generation,
		)
		if err != nil {
			return err
		}
		actual, exists, err := schemaExecutorReadIndex(
			ctx, querier, schemaName, name,
		)
		if err != nil {
			return err
		}
		arguments := schemav2.IndexArguments{
			IndexID: index.IndexID, Generation: index.Generation,
			Fields: index.Fields, Method: index.Method,
			OperatorClass: index.OperatorClass, Unique: index.Unique,
			ActiveRecordsOnly: index.ActiveRecordsOnly,
		}
		if !exists || !actual.valid || !actual.ready ||
			!schemaExecutorIndexMatchesV2(actual, mapping, arguments, trigramSchema) {
			return schemaExecutorBlocked("schema_index_invalid", nil)
		}
	}
	return nil
}
