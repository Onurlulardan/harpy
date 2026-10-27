package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateRelationArgumentsV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.RelationArguments,
) error {
	if _, found := findIdentifierV2(mapping.Relations, arguments.RelationID); !found {
		return fmt.Errorf("relation mapping is missing")
	}
	if !findFormV2(mapping, arguments.TargetFormID) {
		return fmt.Errorf("target form mapping is missing")
	}
	for _, constraintID := range []string{
		arguments.SourceConstraintID, arguments.TargetConstraintID,
	} {
		if _, found := findIdentifierV2(mapping.Constraints, constraintID); !found {
			return fmt.Errorf("relation constraint mapping is missing")
		}
	}
	if arguments.SourceConstraintID == arguments.TargetConstraintID {
		return fmt.Errorf("relation constraints must be distinct")
	}
	if arguments.Cardinality != schemav2.CardinalityOne &&
		arguments.Cardinality != schemav2.CardinalityMany {
		return fmt.Errorf("relation cardinality is invalid")
	}
	if arguments.DeletePolicy != schemav2.DeleteRestrict &&
		arguments.DeletePolicy != schemav2.DeleteCascade {
		return fmt.Errorf("relation delete policy is invalid")
	}
	if arguments.Cardinality == schemav2.CardinalityOne {
		if arguments.CardinalityConstraintID == nil {
			return fmt.Errorf("one-cardinality constraint is missing")
		}
		_, found := findIdentifierV2(mapping.Constraints, *arguments.CardinalityConstraintID)
		if !found || *arguments.CardinalityConstraintID == arguments.SourceConstraintID ||
			*arguments.CardinalityConstraintID == arguments.TargetConstraintID {
			return fmt.Errorf("one-cardinality constraint mapping is invalid")
		}
	} else if arguments.CardinalityConstraintID != nil {
		return fmt.Errorf("many-cardinality cannot carry a unique constraint")
	}
	return nil
}

func findFormV2(mapping schemav2.IdentifierMapping, formID string) bool {
	if mapping.Form.LogicalID == formID {
		return true
	}
	_, found := findIdentifierV2(mapping.RelatedForms, formID)
	return found
}
