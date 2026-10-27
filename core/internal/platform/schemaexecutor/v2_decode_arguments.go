package schemaexecutor

import (
	"encoding/json"
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func decodeArgumentsV2(step schemav2.Step) (any, error) {
	if err := validateArgumentPropertiesV2(step.OperationID, step.Arguments); err != nil {
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
	case schemav2.CreateFormTable, schemav2.DropFormTable:
		if len(step.Arguments) != 0 {
			return nil, fmt.Errorf("%s arguments must be empty", step.OperationID)
		}
		return schemav2.EmptyArguments{}, nil
	case schemav2.AddFieldColumn:
		return decode(&schemav2.ColumnArguments{})
	case schemav2.DropFieldColumn, schemav2.SetFieldNotNull:
		return decode(&schemav2.FieldArguments{})
	case schemav2.AddRelationTable:
		return decode(&schemav2.RelationArguments{})
	case schemav2.DropRelationTable, schemav2.DropIndex,
		schemav2.DropIndexConcurrent, schemav2.DropConstraint:
		return decode(&schemav2.ObjectArguments{})
	case schemav2.CreateIndex, schemav2.CreateIndexConcurrent,
		schemav2.CleanupInvalidIndex:
		return decode(&schemav2.IndexArguments{})
	case schemav2.AddConstraint, schemav2.ValidateConstraint:
		return decode(&schemav2.ConstraintArguments{})
	case schemav2.BackfillFieldBatch:
		return decode(&schemav2.BackfillArguments{})
	case schemav2.VerifyStorage:
		return decode(&schemav2.VerifyArguments{})
	case schemav2.ActivateGeneration, schemav2.DisableOldWriter,
		schemav2.EnableDualWrite:
		return decode(&schemav2.GenerationArguments{})
	case schemav2.RetainGeneration:
		return decode(&schemav2.RetentionArguments{})
	default:
		return nil, fmt.Errorf("operation is not supported")
	}
}
