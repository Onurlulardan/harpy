package schemaexecutor

import (
	"encoding/json"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func decodeArguments(step schemav1.Step) (any, error) {
	if err := validateArgumentProperties(step.OperationID, step.Arguments); err != nil {
		return nil, err
	}
	document, err := json.Marshal(step.Arguments)
	if err != nil {
		return nil, fmt.Errorf("encode operation arguments: %w", err)
	}
	decode := func(target any) (any, error) {
		if err := decodeStrict(document, target); err != nil {
			return nil, fmt.Errorf("%s arguments: %w", step.OperationID, err)
		}
		return target, nil
	}
	switch step.OperationID {
	case schemav1.CreateFormTable, schemav1.DropFormTable:
		if len(step.Arguments) != 0 {
			return nil, fmt.Errorf("%s arguments must be empty", step.OperationID)
		}
		return schemav1.EmptyArguments{}, nil
	case schemav1.AddFieldColumn:
		return decode(&schemav1.ColumnArguments{})
	case schemav1.DropFieldColumn:
		return decode(&schemav1.FieldArguments{})
	case schemav1.AddRelationTable:
		return decode(&schemav1.RelationArguments{})
	case schemav1.DropRelationTable,
		schemav1.DropIndex, schemav1.DropIndexConcurrent,
		schemav1.DropConstraint:
		return decode(&schemav1.ObjectArguments{})
	case schemav1.CreateIndex, schemav1.CreateIndexConcurrent,
		schemav1.CleanupInvalidIndex:
		return decode(&schemav1.IndexArguments{})
	case schemav1.AddConstraint, schemav1.ValidateConstraint:
		return decode(&schemav1.ConstraintArguments{})
	case schemav1.BackfillFieldBatch:
		return decode(&schemav1.BackfillArguments{})
	case schemav1.VerifyStorage:
		return decode(&schemav1.VerifyArguments{})
	case schemav1.ActivateGeneration, schemav1.DisableOldWriter,
		schemav1.EnableDualWrite:
		return decode(&schemav1.GenerationArguments{})
	case schemav1.RetainGeneration:
		return decode(&schemav1.RetentionArguments{})
	default:
		return nil, fmt.Errorf("operation is not supported")
	}
}
