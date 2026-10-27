package postgres

import (
	"strconv"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func schemaExecutorConstraintMatches(
	actual schemaExecutorConstraintContract,
	mapping schemav1.IdentifierMapping,
	expected schemav1.VerifyConstraintContract,
) bool {
	if actual.owner != "harpy_owner" {
		return false
	}
	fields := make([]string, len(expected.Fields))
	for index, field := range expected.Fields {
		name, found := schemaExecutorFieldGenerationName(
			mapping, field.FieldID, field.Generation,
		)
		if !found {
			return false
		}
		fields[index] = name
	}
	definition := schemaExecutorNormalizeDefinition(actual.definition)
	switch expected.Kind {
	case schemav1.ConstraintUnique:
		return actual.kind == "u" && definition == "unique"+strings.Join(fields, ",")
	case schemav1.ConstraintNotNull:
		return actual.kind == "c" && len(fields) == 1 &&
			definition == "check"+fields[0]+"isnotnull"
	case schemav1.ConstraintTextLength:
		return actual.kind == "c" && len(fields) == 1 &&
			schemaExecutorTextLengthMatches(
				definition, fields[0], expected.Minimum, expected.Maximum,
			)
	default:
		return false
	}
}

func schemaExecutorNormalizeDefinition(value string) string {
	replacer := strings.NewReplacer(
		" ", "", "\n", "", "\r", "", "\t", "", "\"", "", "(", "", ")", "",
		"pg_catalog.", "",
	)
	normalized := strings.ToLower(replacer.Replace(value))
	return strings.TrimSuffix(normalized, "notvalid")
}

func schemaExecutorTextLengthMatches(
	definition string,
	field string,
	minimum *int,
	maximum *int,
) bool {
	if minimum == nil || maximum == nil {
		return false
	}
	minValue, maxValue := strconv.Itoa(*minimum), strconv.Itoa(*maximum)
	function := "char_length" + field
	between := "check" + function + "between" + minValue + "and" + maxValue
	bounds := "check" + function + ">=" + minValue + "and" + function + "<=" + maxValue
	return definition == between || definition == bounds
}

func schemaExecutorConstraintArgumentsMatch(
	actual schemaExecutorConstraintContract,
	mapping schemav1.IdentifierMapping,
	expected schemav1.ConstraintArguments,
) bool {
	fields := make([]schemav1.FieldReference, len(expected.FieldIDs))
	for index, fieldID := range expected.FieldIDs {
		field, found := schemaExecutorUniqueField(mapping, fieldID)
		if !found {
			return false
		}
		fields[index] = schemav1.FieldReference{
			FieldID: field.LogicalID, Generation: field.Generation,
		}
	}
	return schemaExecutorConstraintMatches(
		actual, mapping, schemav1.VerifyConstraintContract{
			ConstraintID: expected.ConstraintID, Kind: expected.Kind,
			Fields: fields, Minimum: expected.Minimum, Maximum: expected.Maximum,
		},
	)
}

func schemaExecutorUniqueField(
	mapping schemav1.IdentifierMapping,
	fieldID string,
) (schemav1.FieldIdentifier, bool) {
	var result schemav1.FieldIdentifier
	count := 0
	for _, field := range mapping.Fields {
		if field.LogicalID == fieldID {
			result, count = field, count+1
		}
	}
	return result, count == 1
}
