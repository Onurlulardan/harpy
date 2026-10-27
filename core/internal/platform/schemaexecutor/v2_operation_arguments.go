package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateOperationArgumentsV2(step schemav2.Step, value any) error {
	mapping := step.IdentifierMapping
	switch arguments := value.(type) {
	case schemav2.EmptyArguments:
		return nil
	case *schemav2.ColumnArguments:
		return validateColumnArgumentsV2(mapping, *arguments)
	case *schemav2.FieldArguments:
		return requireFieldV2(mapping, schemav2.FieldReference{
			FieldID: arguments.FieldID, Generation: arguments.Generation,
			Variant: arguments.Variant,
		})
	case *schemav2.RelationArguments:
		return validateRelationArgumentsV2(mapping, *arguments)
	case *schemav2.ObjectArguments:
		return validateObjectArgumentsV2(step.OperationID, mapping, *arguments)
	case *schemav2.IndexArguments:
		return validateIndexArgumentsV2(mapping, *arguments)
	case *schemav2.ConstraintArguments:
		return validateConstraintArgumentsV2(mapping, *arguments)
	case *schemav2.BackfillArguments:
		return validateBackfillArgumentsV2(mapping, *arguments)
	case *schemav2.VerifyArguments:
		return validateVerifyArgumentsV2(mapping, *arguments)
	case *schemav2.GenerationArguments:
		return validateGenerationArgumentsV2(*arguments)
	case *schemav2.RetentionArguments:
		return validateRetentionArgumentsV2(*arguments)
	default:
		return fmt.Errorf("operation argument type is not registered")
	}
}
