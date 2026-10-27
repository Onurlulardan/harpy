package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateConstraintArgumentsV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.ConstraintArguments,
) error {
	if arguments.Generation < 1 ||
		arguments.Generation > schemav2.MaximumFieldGeneration {
		return fmt.Errorf("constraint generation is invalid")
	}
	if _, found := findConstraintIdentifierV2(
		mapping.Constraints, arguments.ConstraintID, arguments.Generation,
	); !found {
		return fmt.Errorf("constraint mapping is missing")
	}
	if len(arguments.Fields) < 1 || len(arguments.Fields) > 16 {
		return fmt.Errorf("constraint field count is invalid")
	}
	seen := map[schemav2.FieldReference]struct{}{}
	for _, field := range arguments.Fields {
		if field.Variant != schemav2.VariantValue {
			return fmt.Errorf("constraint requires value field")
		}
		if err := requireFieldV2(mapping, field); err != nil {
			return err
		}
		if _, exists := seen[field]; exists {
			return fmt.Errorf("constraint field is duplicated")
		}
		seen[field] = struct{}{}
	}
	switch arguments.Kind {
	case schemav2.ConstraintNotNull:
		if len(arguments.Fields) != 1 || arguments.Minimum != nil || arguments.Maximum != nil {
			return fmt.Errorf("not-null constraint shape is invalid")
		}
	case schemav2.ConstraintUnique:
		if arguments.Minimum != nil || arguments.Maximum != nil {
			return fmt.Errorf("unique constraint bounds are forbidden")
		}
	case schemav2.ConstraintTextLength:
		if len(arguments.Fields) != 1 || arguments.Minimum == nil ||
			arguments.Maximum == nil || *arguments.Minimum < 0 ||
			*arguments.Maximum < *arguments.Minimum || *arguments.Maximum > 1048576 {
			return fmt.Errorf("text-length constraint bounds are invalid")
		}
	default:
		return fmt.Errorf("constraint kind is invalid")
	}
	return nil
}
