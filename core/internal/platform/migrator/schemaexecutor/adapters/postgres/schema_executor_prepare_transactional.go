package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorPrepareTransactional(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	trigramSchema string,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	switch step.Operation() {
	case schemav1.CreateFormTable, schemav1.DropFormTable,
		schemav1.AddFieldColumn, schemav1.DropFieldColumn,
		schemav1.OperationID(schemav2.SetFieldNotNull):
		return schemaExecutorPrepareForm(ctx, querier, schemaName, step)
	case schemav1.AddRelationTable, schemav1.DropRelationTable:
		return schemaExecutorPrepareRelation(ctx, querier, schemaName, step)
	case schemav1.AddConstraint, schemav1.DropConstraint,
		schemav1.ValidateConstraint:
		return schemaExecutorPrepareConstraint(ctx, querier, schemaName, step)
	case schemav1.CreateIndex, schemav1.DropIndex:
		prepared, err := schemaExecutorPrepareIndex(
			ctx, querier, schemaName, trigramSchema, step,
		)
		return prepared.apply, err
	case schemav1.BackfillFieldBatch:
		if mapping, ok := step.V2Mapping(); ok {
			return false, schemaExecutorRequireTable(
				ctx, querier, schemaName, mapping.Form.PhysicalName,
			)
		}
		return false, schemaExecutorRequireTable(
			ctx, querier, schemaName, step.Mapping().Form.PhysicalName,
		)
	case schemav1.VerifyStorage:
		return false, nil
	default:
		return false, fmt.Errorf("transactional operation is not registered")
	}
}
