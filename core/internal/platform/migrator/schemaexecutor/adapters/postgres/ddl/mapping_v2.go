package ddl

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func fieldNameV2(
	mapping schemav2.IdentifierMapping,
	reference schemav2.FieldReference,
) (string, error) {
	for _, field := range mapping.Fields {
		if field.LogicalID == reference.FieldID && field.Generation == reference.Generation &&
			field.Variant == reference.Variant {
			return field.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated V2 field mapping is missing")
}

func identifierNameV2(values []schemav2.Identifier, logicalID string) (string, error) {
	for _, value := range values {
		if value.LogicalID == logicalID {
			return value.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated V2 object mapping is missing")
}

func indexNameV2(
	values []schemav2.Identifier, logicalID string, generation int64,
) (string, error) {
	for _, value := range values {
		if value.LogicalID == logicalID && value.Generation == generation {
			return value.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated V2 index mapping is missing")
}

func constraintNameV2(
	values []schemav2.Identifier, logicalID string, generation int64,
) (string, error) {
	for _, value := range values {
		if value.LogicalID == logicalID && value.Generation == generation {
			return value.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated V2 constraint mapping is missing")
}

func formNameV2(mapping schemav2.IdentifierMapping, logicalID string) (string, error) {
	if mapping.Form.LogicalID == logicalID {
		return mapping.Form.PhysicalName, nil
	}
	return identifierNameV2(mapping.RelatedForms, logicalID)
}

func v2Mapping(stepMapping func() (schemav2.IdentifierMapping, bool)) (
	schemav2.IdentifierMapping,
	error,
) {
	mapping, ok := stepMapping()
	if !ok {
		return schemav2.IdentifierMapping{}, fmt.Errorf("validated V2 mapping is unavailable")
	}
	return mapping, nil
}
