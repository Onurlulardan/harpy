package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateOperationArguments(step schemav1.Step, value any) error {
	mapping := step.IdentifierMapping
	switch arguments := value.(type) {
	case schemav1.EmptyArguments:
		if step.OperationID == schemav1.DropFormTable {
			return nil
		}
		return nil
	case *schemav1.ColumnArguments:
		return validateColumnArguments(mapping, *arguments)
	case *schemav1.FieldArguments:
		return requireField(mapping, arguments.FieldID)
	case *schemav1.RelationArguments:
		return validateRelationArguments(mapping, *arguments)
	case *schemav1.ObjectArguments:
		return validateObjectArguments(step.OperationID, mapping, arguments.ObjectID)
	case *schemav1.IndexArguments:
		return validateIndexArguments(mapping, *arguments)
	case *schemav1.ConstraintArguments:
		return validateConstraintArguments(mapping, *arguments)
	case *schemav1.BackfillArguments:
		return validateBackfillArguments(mapping, *arguments)
	case *schemav1.VerifyArguments:
		return validateVerifyArguments(mapping, *arguments)
	case *schemav1.GenerationArguments:
		return validateGenerationArguments(*arguments)
	case *schemav1.RetentionArguments:
		return validateRetentionArguments(*arguments)
	default:
		return fmt.Errorf("operation argument type is not registered")
	}
}

func requireField(mapping schemav1.IdentifierMapping, fieldID string) error {
	if _, err := canonicalUUID(fieldID); err != nil {
		return fmt.Errorf("field ID is invalid")
	}
	if _, found := findField(mapping.Fields, fieldID); !found {
		return fmt.Errorf("field mapping is missing or ambiguous")
	}
	return nil
}
