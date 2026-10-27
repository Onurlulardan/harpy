package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorPrepareRelationV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	mapping, err := schemaExecutorMappingV2(step)
	if err != nil {
		return false, err
	}
	if step.Operation() == schemav1.DropRelationTable {
		arguments, ok := step.ArgumentValue().(*schemav2.ObjectArguments)
		if !ok {
			return false, fmt.Errorf("V2 relation drop arguments are invalid")
		}
		name, err := schemaExecutorIdentifierNameV2(mapping.Relations, arguments.ObjectID)
		if err != nil {
			return false, err
		}
		return schemaExecutorOwnedTable(ctx, querier, schemaName, name)
	}
	arguments, ok := step.ArgumentValue().(*schemav2.RelationArguments)
	if !ok {
		return false, fmt.Errorf("V2 relation add arguments are invalid")
	}
	if err := schemaExecutorRequireTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	); err != nil {
		return false, err
	}
	target, err := schemaExecutorFormNameV2(mapping, arguments.TargetFormID)
	if err != nil {
		return false, err
	}
	if err := schemaExecutorRequireTable(ctx, querier, schemaName, target); err != nil {
		return false, err
	}
	relation, err := schemaExecutorIdentifierNameV2(
		mapping.Relations, arguments.RelationID,
	)
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
