package schemaexecutor

import schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"

func findIdentifier(values []schemav1.Identifier, logicalID string) (schemav1.Identifier, bool) {
	for _, value := range values {
		if value.LogicalID == logicalID {
			return value, true
		}
	}
	return schemav1.Identifier{}, false
}

func findField(
	values []schemav1.FieldIdentifier,
	logicalID string,
) (schemav1.FieldIdentifier, bool) {
	var found schemav1.FieldIdentifier
	matches := 0
	for _, value := range values {
		if value.LogicalID == logicalID {
			found = value
			matches++
		}
	}
	return found, matches == 1
}

func findFieldGeneration(
	values []schemav1.FieldIdentifier,
	logicalID string,
	generation int64,
) (schemav1.FieldIdentifier, bool) {
	for _, value := range values {
		if value.LogicalID == logicalID && value.Generation == generation {
			return value, true
		}
	}
	return schemav1.FieldIdentifier{}, false
}

func findForm(mapping schemav1.IdentifierMapping, logicalID string) (schemav1.Identifier, bool) {
	if mapping.Form.LogicalID == logicalID {
		return mapping.Form, true
	}
	return findIdentifier(mapping.RelatedForms, logicalID)
}
