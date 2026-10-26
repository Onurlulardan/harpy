package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorReconcileIndex(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	trigramSchema string,
	step schemaexecutor.ValidatedStep,
) error {
	prepared, err := schemaExecutorPrepareIndex(
		ctx, querier, schemaName, trigramSchema, step,
	)
	if err != nil {
		return err
	}
	switch step.Operation() {
	case schemav1.CreateIndexConcurrent:
		if !prepared.exists || !prepared.contract.valid || !prepared.contract.ready {
			return schemaExecutorBlocked("schema_index_invalid", nil)
		}
	case schemav1.CleanupInvalidIndex, schemav1.DropIndexConcurrent:
		if prepared.exists {
			return schemaExecutorBlocked("schema_index_cleanup_incomplete", nil)
		}
	default:
		return fmt.Errorf("concurrent index operation is not registered")
	}
	return nil
}

func schemaExecutorCleanupStatement(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (ddl.Statement, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		mapping, err := schemaExecutorMappingV2(step)
		if err != nil {
			return ddl.Statement{}, err
		}
		arguments, ok := step.ArgumentValue().(*schemav2.IndexArguments)
		if !ok {
			return ddl.Statement{}, fmt.Errorf("V2 cleanup arguments are invalid")
		}
		name, err := schemaExecutorIndexNameV2(
			mapping.Indexes, arguments.IndexID, arguments.Generation,
		)
		if err != nil {
			return ddl.Statement{}, err
		}
		return ddl.DropIndexConcurrently(schemaName, name), nil
	}
	arguments, ok := step.ArgumentValue().(*schemav1.IndexArguments)
	if !ok {
		return ddl.Statement{}, fmt.Errorf("index cleanup arguments are invalid")
	}
	mapping := step.Mapping()
	name, err := schemaExecutorIdentifierName(mapping.Indexes, arguments.IndexID)
	if err != nil {
		return ddl.Statement{}, err
	}
	return ddl.DropIndexConcurrently(schemaName, name), nil
}
