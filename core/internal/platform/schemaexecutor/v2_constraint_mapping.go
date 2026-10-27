package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateConstraintIdentifiersV2(
	values []schemav2.Identifier, seenNames map[string]struct{},
) error {
	previous := ""
	for _, value := range values {
		expected, err := constraintIdentifierNameV2(value)
		key := value.LogicalID + ":" + fmt.Sprintf("%010d", value.Generation)
		if err != nil || value.PhysicalName != expected {
			return fmt.Errorf("constraint physical identifier is not deterministic")
		}
		if previous != "" && previous >= key {
			return fmt.Errorf("constraint identifiers are not sorted and unique")
		}
		if _, exists := seenNames[value.PhysicalName]; exists {
			return fmt.Errorf("physical identifier is duplicated")
		}
		seenNames[value.PhysicalName] = struct{}{}
		previous = key
	}
	return nil
}

func constraintIdentifierNameV2(value schemav2.Identifier) (string, error) {
	if value.Generation == 0 {
		return schemav2.DeriveIdentifier(schemav2.IdentifierConstraint, value.LogicalID)
	}
	return schemav2.DeriveConstraintIdentifier(value.LogicalID, value.Generation)
}

func findConstraintIdentifierV2(
	values []schemav2.Identifier, logicalID string, generation int64,
) (schemav2.Identifier, bool) {
	for _, value := range values {
		if value.LogicalID == logicalID && value.Generation == generation {
			return value, true
		}
	}
	return schemav2.Identifier{}, false
}
