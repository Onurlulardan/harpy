package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateObjectArgumentsV2(
	operation schemav2.OperationID,
	mapping schemav2.IdentifierMapping,
	arguments schemav2.ObjectArguments,
) error {
	var values []schemav2.Identifier
	switch operation {
	case schemav2.DropRelationTable:
		values = mapping.Relations
	case schemav2.DropIndex, schemav2.DropIndexConcurrent:
		if arguments.Generation == nil {
			return fmt.Errorf("index generation is required")
		}
		if _, found := findIndexIdentifierV2(
			mapping.Indexes, arguments.ObjectID, *arguments.Generation,
		); !found {
			return fmt.Errorf("object mapping is missing")
		}
		return nil
	case schemav2.DropConstraint:
		if arguments.Generation == nil {
			return fmt.Errorf("constraint generation is required")
		}
		if _, found := findConstraintIdentifierV2(
			mapping.Constraints, arguments.ObjectID, *arguments.Generation,
		); !found {
			return fmt.Errorf("object mapping is missing")
		}
		return nil
	default:
		return fmt.Errorf("object operation is invalid")
	}
	if arguments.Generation != nil {
		return fmt.Errorf("object generation is forbidden")
	}
	if _, found := findIdentifierV2(values, arguments.ObjectID); !found {
		return fmt.Errorf("object mapping is missing")
	}
	return nil
}
