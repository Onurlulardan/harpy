package schemaexecutor

import (
	"fmt"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateFields(
	fields []schemav1.FieldIdentifier,
	seenNames map[string]struct{},
) error {
	previousID := ""
	previousGeneration := int64(0)
	for _, field := range fields {
		expected, err := deriveFieldIdentifier(field.LogicalID, field.Generation)
		if err != nil || field.PhysicalName != expected {
			return fmt.Errorf("field physical identifier is not deterministic")
		}
		if strings.Compare(previousID, field.LogicalID) > 0 ||
			(previousID == field.LogicalID && previousGeneration >= field.Generation) {
			return fmt.Errorf("field identifiers are not sorted and unique")
		}
		if _, exists := seenNames[field.PhysicalName]; exists {
			return fmt.Errorf("field physical identifier is duplicated")
		}
		seenNames[field.PhysicalName] = struct{}{}
		previousID, previousGeneration = field.LogicalID, field.Generation
	}
	return nil
}
