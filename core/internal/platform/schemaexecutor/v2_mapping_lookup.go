package schemaexecutor

import schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"

func findFieldV2(
	values []schemav2.FieldIdentifier,
	reference schemav2.FieldReference,
) (schemav2.FieldIdentifier, bool) {
	var result schemav2.FieldIdentifier
	found := false
	for _, value := range values {
		if value.LogicalID == reference.FieldID && value.Generation == reference.Generation &&
			value.Variant == reference.Variant {
			if found {
				return schemav2.FieldIdentifier{}, false
			}
			result, found = value, true
		}
	}
	return result, found
}

func requireFieldV2(
	mapping schemav2.IdentifierMapping,
	reference schemav2.FieldReference,
) error {
	if _, err := canonicalUUID(reference.FieldID); err != nil ||
		reference.Generation < 1 || reference.Generation > schemav2.MaximumFieldGeneration {
		return errInvalidFieldReference
	}
	if _, found := findFieldV2(mapping.Fields, reference); !found {
		return errMissingFieldReference
	}
	return nil
}

var (
	errInvalidFieldReference = contractError("field reference is invalid")
	errMissingFieldReference = contractError("field mapping is missing or ambiguous")
)

type contractError string

func (value contractError) Error() string { return string(value) }

func indexReference(field schemav2.IndexField) schemav2.FieldReference {
	return schemav2.FieldReference{
		FieldID: field.FieldID, Generation: field.Generation, Variant: field.Variant,
	}
}
