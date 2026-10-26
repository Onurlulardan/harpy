package postgres

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorMappingV2(
	step schemaexecutor.ValidatedStep,
) (schemav2.IdentifierMapping, error) {
	mapping, ok := step.V2Mapping()
	if !ok {
		return schemav2.IdentifierMapping{}, fmt.Errorf("validated V2 mapping is unavailable")
	}
	return mapping, nil
}

func schemaExecutorFieldNameV2(
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

func schemaExecutorIdentifierNameV2(
	values []schemav2.Identifier,
	logicalID string,
) (string, error) {
	for _, value := range values {
		if value.LogicalID == logicalID {
			return value.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated V2 object mapping is missing")
}

func schemaExecutorIndexNameV2(
	values []schemav2.Identifier, logicalID string, generation int64,
) (string, error) {
	for _, value := range values {
		if value.LogicalID == logicalID && value.Generation == generation {
			return value.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated V2 index mapping is missing")
}

func schemaExecutorConstraintNameV2(
	values []schemav2.Identifier, logicalID string, generation int64,
) (string, error) {
	for _, value := range values {
		if value.LogicalID == logicalID && value.Generation == generation {
			return value.PhysicalName, nil
		}
	}
	return "", fmt.Errorf("validated V2 constraint mapping is missing")
}

func schemaExecutorFormNameV2(
	mapping schemav2.IdentifierMapping,
	logicalID string,
) (string, error) {
	if mapping.Form.LogicalID == logicalID {
		return mapping.Form.PhysicalName, nil
	}
	return schemaExecutorIdentifierNameV2(mapping.RelatedForms, logicalID)
}
