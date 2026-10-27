package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

type schemaExecutorIndexPreparation struct {
	apply    bool
	exists   bool
	contract schemaExecutorIndexContract
}

func schemaExecutorPrepareIndex(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	trigramSchema string,
	step schemaexecutor.ValidatedStep,
) (schemaExecutorIndexPreparation, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return schemaExecutorPrepareIndexV2(
			ctx, querier, schemaName, trigramSchema, step,
		)
	}
	mapping := step.Mapping()
	if err := schemaExecutorRequireTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	); err != nil {
		return schemaExecutorIndexPreparation{}, err
	}
	var indexID string
	var expected *schemav1.IndexArguments
	if step.Operation() == schemav1.CreateIndex ||
		step.Operation() == schemav1.CreateIndexConcurrent ||
		step.Operation() == schemav1.CleanupInvalidIndex {
		expected = step.ArgumentValue().(*schemav1.IndexArguments)
		indexID = expected.IndexID
	} else {
		indexID = step.ArgumentValue().(*schemav1.ObjectArguments).ObjectID
	}
	name, err := schemaExecutorIdentifierName(mapping.Indexes, indexID)
	if err != nil {
		return schemaExecutorIndexPreparation{}, err
	}
	contract, exists, err := schemaExecutorReadIndex(
		ctx, querier, schemaName, name,
	)
	if err != nil {
		return schemaExecutorIndexPreparation{}, err
	}
	result := schemaExecutorIndexPreparation{exists: exists, contract: contract}
	if expected != nil {
		if step.Operation() == schemav1.CleanupInvalidIndex {
			if !exists {
				return result, nil
			}
			if !schemaExecutorIndexMatches(contract, mapping, *expected) {
				return result, schemaExecutorBlocked("schema_index_contract_mismatch", nil)
			}
			result.apply = !contract.valid || !contract.ready
			return result, nil
		}
		if !exists {
			result.apply = true
			return result, nil
		}
		if !schemaExecutorIndexMatches(contract, mapping, *expected) {
			return result, schemaExecutorBlocked("schema_index_contract_mismatch", nil)
		}
		if step.Operation() == schemav1.CreateIndex &&
			(!contract.valid || !contract.ready) {
			return result, schemaExecutorBlocked("schema_index_invalid", nil)
		}
		return result, nil
	}
	if exists && contract.owner != "harpy_owner" {
		return result, schemaExecutorBlocked("schema_object_ownership_invalid", nil)
	}
	if exists && contract.tableName != mapping.Form.PhysicalName {
		return result, schemaExecutorBlocked("schema_index_contract_mismatch", nil)
	}
	result.apply = exists
	return result, nil
}

func schemaExecutorIndexArguments(
	step schemaexecutor.ValidatedStep,
) (schemav1.IndexArguments, error) {
	arguments, ok := step.ArgumentValue().(*schemav1.IndexArguments)
	if !ok {
		return schemav1.IndexArguments{}, fmt.Errorf("index arguments are invalid")
	}
	return *arguments, nil
}
