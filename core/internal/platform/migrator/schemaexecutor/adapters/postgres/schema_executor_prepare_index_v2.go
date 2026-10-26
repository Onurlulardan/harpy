package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorPrepareIndexV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	trigramSchema string,
	step schemaexecutor.ValidatedStep,
) (schemaExecutorIndexPreparation, error) {
	mapping, err := schemaExecutorMappingV2(step)
	if err != nil {
		return schemaExecutorIndexPreparation{}, err
	}
	if err = schemaExecutorRequireTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	); err != nil {
		return schemaExecutorIndexPreparation{}, err
	}
	indexID, generation, expected, err := schemaExecutorIndexIdentityV2(step)
	if err != nil {
		return schemaExecutorIndexPreparation{}, err
	}
	name, err := schemaExecutorIndexNameV2(mapping.Indexes, indexID, generation)
	if err != nil {
		return schemaExecutorIndexPreparation{}, err
	}
	contract, exists, err := schemaExecutorReadIndex(ctx, querier, schemaName, name)
	if err != nil {
		return schemaExecutorIndexPreparation{}, err
	}
	result := schemaExecutorIndexPreparation{exists: exists, contract: contract}
	if expected == nil {
		if exists && (contract.owner != "harpy_owner" ||
			contract.tableName != mapping.Form.PhysicalName) {
			return result, schemaExecutorBlocked("schema_index_contract_mismatch", nil)
		}
		result.apply = exists
		return result, nil
	}
	if !exists {
		result.apply = step.Operation() != schemav1.CleanupInvalidIndex
		return result, nil
	}
	if !schemaExecutorIndexMatchesV2(
		contract, mapping, *expected, trigramSchema,
	) {
		return result, schemaExecutorBlocked("schema_index_contract_mismatch", nil)
	}
	if step.Operation() == schemav1.CleanupInvalidIndex {
		result.apply = !contract.valid || !contract.ready
		return result, nil
	}
	if step.Operation() == schemav1.CreateIndex &&
		(!contract.valid || !contract.ready) {
		return result, schemaExecutorBlocked("schema_index_invalid", nil)
	}
	return result, nil
}

func schemaExecutorIndexIdentityV2(
	step schemaexecutor.ValidatedStep,
) (string, int64, *schemav2.IndexArguments, error) {
	switch step.Operation() {
	case schemav1.CreateIndex, schemav1.CreateIndexConcurrent,
		schemav1.CleanupInvalidIndex:
		arguments, ok := step.ArgumentValue().(*schemav2.IndexArguments)
		if !ok {
			return "", 0, nil, fmt.Errorf("V2 index arguments are invalid")
		}
		return arguments.IndexID, arguments.Generation, arguments, nil
	default:
		arguments, ok := step.ArgumentValue().(*schemav2.ObjectArguments)
		if !ok {
			return "", 0, nil, fmt.Errorf("V2 index object arguments are invalid")
		}
		if arguments.Generation == nil {
			return "", 0, nil, fmt.Errorf("V2 index generation is missing")
		}
		return arguments.ObjectID, *arguments.Generation, nil, nil
	}
}
