package postgres

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func schemaExecutorFieldName(
	mapping schemav1.IdentifierMapping,
	logicalID string,
) (string, error) {
	for _, field := range mapping.Fields {
		if field.LogicalID == logicalID {
			return field.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated field mapping is missing")
}

func schemaExecutorIdentifierName(
	values []schemav1.Identifier,
	logicalID string,
) (string, error) {
	for _, value := range values {
		if value.LogicalID == logicalID {
			return value.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated object mapping is missing")
}

func schemaExecutorFormName(
	mapping schemav1.IdentifierMapping,
	logicalID string,
) (string, error) {
	if mapping.Form.LogicalID == logicalID {
		return mapping.Form.PhysicalName, nil
	}
	return schemaExecutorIdentifierName(mapping.RelatedForms, logicalID)
}
