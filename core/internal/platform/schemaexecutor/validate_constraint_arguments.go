package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateConstraintArguments(
	mapping schemav1.IdentifierMapping,
	arguments schemav1.ConstraintArguments,
) error {
	if _, found := findIdentifier(mapping.Constraints, arguments.ConstraintID); !found {
		return fmt.Errorf("constraint mapping is missing")
	}
	if len(arguments.FieldIDs) < 1 || len(arguments.FieldIDs) > 16 {
		return fmt.Errorf("constraint field count is invalid")
	}
	seen := map[string]struct{}{}
	for _, fieldID := range arguments.FieldIDs {
		if err := requireField(mapping, fieldID); err != nil {
			return err
		}
		if _, exists := seen[fieldID]; exists {
			return fmt.Errorf("constraint field is duplicated")
		}
		seen[fieldID] = struct{}{}
	}
	switch arguments.Kind {
	case schemav1.ConstraintNotNull:
		if len(arguments.FieldIDs) != 1 || arguments.Minimum != nil || arguments.Maximum != nil {
			return fmt.Errorf("not-null constraint shape is invalid")
		}
	case schemav1.ConstraintUnique:
		if arguments.Minimum != nil || arguments.Maximum != nil {
			return fmt.Errorf("unique constraint bounds are forbidden")
		}
	case schemav1.ConstraintTextLength:
		if len(arguments.FieldIDs) != 1 || arguments.Minimum == nil ||
			arguments.Maximum == nil || *arguments.Minimum < 0 ||
			*arguments.Maximum < *arguments.Minimum || *arguments.Maximum > 1048576 {
			return fmt.Errorf("text-length constraint bounds are invalid")
		}
	default:
		return fmt.Errorf("constraint kind is invalid")
	}
	return nil
}
