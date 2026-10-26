package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorPrepareRelation(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return schemaExecutorPrepareRelationV2(ctx, querier, schemaName, step)
	}
	mapping := step.Mapping()
	if step.Operation() == schemav1.DropRelationTable {
		arguments := step.ArgumentValue().(*schemav1.ObjectArguments)
		relation, err := schemaExecutorIdentifierName(mapping.Relations, arguments.ObjectID)
		if err != nil {
			return false, err
		}
		return schemaExecutorOwnedTable(ctx, querier, schemaName, relation)
	}
	arguments, ok := step.ArgumentValue().(*schemav1.RelationArguments)
	if !ok {
		return false, fmt.Errorf("relation operation arguments are invalid")
	}
	if err := schemaExecutorRequireTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	); err != nil {
		return false, err
	}
	target, err := schemaExecutorFormName(mapping, arguments.TargetFormID)
	if err != nil {
		return false, err
	}
	if err := schemaExecutorRequireTable(ctx, querier, schemaName, target); err != nil {
		return false, err
	}
	relation, err := schemaExecutorIdentifierName(mapping.Relations, arguments.RelationID)
	if err != nil {
		return false, err
	}
	exists, err := schemaExecutorOwnedTable(ctx, querier, schemaName, relation)
	if err != nil {
		return false, err
	}
	if exists {
		return false, schemaExecutorBlocked("schema_object_already_exists", nil)
	}
	return true, nil
}
