package postgres

import (
	"context"
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorVerifyStorageV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	trigramSchema string,
	step schemaexecutor.ValidatedStep,
) error {
	arguments, ok := step.ArgumentValue().(*schemav2.VerifyArguments)
	if !ok {
		return fmt.Errorf("V2 verify arguments are invalid")
	}
	mapping, err := schemaExecutorMappingV2(step)
	if err != nil {
		return err
	}
	for _, check := range arguments.Checks {
		switch check {
		case schemav2.VerifyFormTable:
			err = schemaExecutorVerifyFormV2(
				ctx, querier, schemaName, mapping, *arguments,
			)
		case schemav2.VerifySystemColumns:
			continue
		case schemav2.VerifyFields:
			err = schemaExecutorVerifyFieldsV2(
				ctx, querier, schemaName, mapping, arguments.Fields,
			)
		case schemav2.VerifyRelations:
			err = schemaExecutorVerifyRelationsV2(
				ctx, querier, schemaName, mapping, arguments.Relations,
			)
		case schemav2.VerifyIndexes:
			err = schemaExecutorVerifyIndexesV2(
				ctx, querier, schemaName, trigramSchema, mapping, arguments.Indexes,
			)
		case schemav2.VerifyConstraints:
			err = schemaExecutorVerifyConstraintsV2(
				ctx, querier, schemaName, mapping, arguments.Constraints,
			)
		case schemav2.VerifyEvolutionData:
			err = schemaExecutorVerifyEvolutionDataV2(
				ctx, querier, schemaName, mapping, arguments.Evolution,
			)
		default:
			return schemaExecutorBlocked("schema_verify_check_unknown", nil)
		}
		if err != nil {
			return err
		}
	}
	return schemaExecutorVerifyRecordACLsV2(ctx, querier, schemaName, mapping)
}
