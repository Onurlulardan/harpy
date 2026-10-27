package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateRelationArguments(
	mapping schemav1.IdentifierMapping,
	arguments schemav1.RelationArguments,
) error {
	if _, found := findIdentifier(mapping.Relations, arguments.RelationID); !found {
		return fmt.Errorf("relation mapping is missing")
	}
	if _, found := findForm(mapping, arguments.TargetFormID); !found {
		return fmt.Errorf("target form mapping is missing")
	}
	for _, constraintID := range []string{
		arguments.SourceConstraintID, arguments.TargetConstraintID,
	} {
		if _, found := findIdentifier(mapping.Constraints, constraintID); !found {
			return fmt.Errorf("relation constraint mapping is missing")
		}
	}
	if arguments.SourceConstraintID == arguments.TargetConstraintID {
		return fmt.Errorf("relation constraints must be distinct")
	}
	if arguments.Cardinality != schemav1.CardinalityOne &&
		arguments.Cardinality != schemav1.CardinalityMany {
		return fmt.Errorf("relation cardinality is invalid")
	}
	if arguments.DeletePolicy != schemav1.DeleteRestrict &&
		arguments.DeletePolicy != schemav1.DeleteCascade {
		return fmt.Errorf("relation delete policy is invalid")
	}
	if arguments.Cardinality == schemav1.CardinalityOne {
		if arguments.CardinalityConstraintID == nil {
			return fmt.Errorf("one-cardinality constraint is missing")
		}
		if _, found := findIdentifier(
			mapping.Constraints, *arguments.CardinalityConstraintID,
		); !found || *arguments.CardinalityConstraintID == arguments.SourceConstraintID ||
			*arguments.CardinalityConstraintID == arguments.TargetConstraintID {
			return fmt.Errorf("one-cardinality constraint mapping is invalid")
		}
	} else if arguments.CardinalityConstraintID != nil {
		return fmt.Errorf("many-cardinality cannot carry a unique constraint")
	}
	return nil
}
