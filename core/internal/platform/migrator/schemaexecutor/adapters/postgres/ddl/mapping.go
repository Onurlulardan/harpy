package ddl

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/jackc/pgx/v5"
)

func qualify(schemaName, objectName string) string {
	return pgx.Identifier{schemaName, objectName}.Sanitize()
}

func quote(name string) string {
	return pgx.Identifier{name}.Sanitize()
}

func fieldName(mapping schemav1.IdentifierMapping, logicalID string) (string, error) {
	for _, field := range mapping.Fields {
		if field.LogicalID == logicalID {
			return field.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated field mapping is missing")
}

func fieldGenerationName(
	mapping schemav1.IdentifierMapping,
	logicalID string,
	generation int64,
) (string, error) {
	for _, field := range mapping.Fields {
		if field.LogicalID == logicalID && field.Generation == generation {
			return field.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated field generation mapping is missing")
}

func identifierName(values []schemav1.Identifier, logicalID string) (string, error) {
	for _, value := range values {
		if value.LogicalID == logicalID {
			return value.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated object mapping is missing")
}

func formName(mapping schemav1.IdentifierMapping, logicalID string) (string, error) {
	if mapping.Form.LogicalID == logicalID {
		return mapping.Form.PhysicalName, nil
	}
	return identifierName(mapping.RelatedForms, logicalID)
}
