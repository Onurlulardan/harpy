package postgres

import (
	"strings"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorConstraintArgumentsMatchV2(
	actual schemaExecutorConstraintContract,
	mapping schemav2.IdentifierMapping,
	expected schemav2.ConstraintArguments,
) bool {
	return schemaExecutorConstraintMatchV2(
		actual, mapping, expected.Kind, expected.Fields,
		expected.Minimum, expected.Maximum,
	)
}

func schemaExecutorVerifyConstraintMatchV2(
	actual schemaExecutorConstraintContract,
	mapping schemav2.IdentifierMapping,
	expected schemav2.VerifyConstraintContract,
) bool {
	return schemaExecutorConstraintMatchV2(
		actual, mapping, expected.Kind, expected.Fields,
		expected.Minimum, expected.Maximum,
	)
}

func schemaExecutorConstraintMatchV2(
	actual schemaExecutorConstraintContract,
	mapping schemav2.IdentifierMapping,
	kind schemav2.ConstraintKind,
	fields []schemav2.FieldReference,
	minimum *int,
	maximum *int,
) bool {
	if actual.owner != "harpy_owner" {
		return false
	}
	names := make([]string, len(fields))
	for index, field := range fields {
		name, err := schemaExecutorFieldNameV2(mapping, field)
		if err != nil {
			return false
		}
		names[index] = name
	}
	definition := schemaExecutorNormalizeDefinition(actual.definition)
	switch kind {
	case schemav2.ConstraintUnique:
		return actual.kind == "u" && definition == "unique"+strings.Join(names, ",")
	case schemav2.ConstraintNotNull:
		return actual.kind == "c" && len(names) == 1 &&
			definition == "check"+names[0]+"isnotnull"
	case schemav2.ConstraintTextLength:
		return actual.kind == "c" && len(names) == 1 &&
			schemaExecutorTextLengthMatches(definition, names[0], minimum, maximum)
	default:
		return false
	}
}
