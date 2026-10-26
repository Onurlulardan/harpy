package ddl

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

type RenderContext struct {
	TenantSchema  string
	TrigramSchema string
}

func Render(renderContext RenderContext, step schemaexecutor.ValidatedStep) (Statement, error) {
	schemaName := renderContext.TenantSchema
	switch step.Operation() {
	case schemav1.CreateFormTable, schemav1.DropFormTable:
		return renderFormTable(schemaName, step)
	case schemav1.AddFieldColumn, schemav1.DropFieldColumn,
		schemav1.OperationID(schemav2.SetFieldNotNull):
		return renderFieldColumn(schemaName, step)
	case schemav1.AddRelationTable, schemav1.DropRelationTable:
		return renderRelationTable(schemaName, step)
	case schemav1.CreateIndex, schemav1.CreateIndexConcurrent,
		schemav1.CleanupInvalidIndex, schemav1.DropIndex,
		schemav1.DropIndexConcurrent:
		return renderIndex(renderContext, step)
	case schemav1.AddConstraint, schemav1.DropConstraint,
		schemav1.ValidateConstraint:
		return renderConstraint(schemaName, step)
	case schemav1.BackfillFieldBatch:
		return renderBackfill(schemaName, step)
	default:
		return Statement{}, fmt.Errorf("operation has no DDL renderer")
	}
}
