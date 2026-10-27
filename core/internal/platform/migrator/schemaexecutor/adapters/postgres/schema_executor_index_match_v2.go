package postgres

import (
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorIndexMatchesV2(
	actual schemaExecutorIndexContract,
	mapping schemav2.IdentifierMapping,
	expected schemav2.IndexArguments,
	trigramSchema string,
) bool {
	fields := make([]string, len(expected.Fields))
	for index, field := range expected.Fields {
		name, err := schemaExecutorFieldNameV2(mapping, schemav2.FieldReference{
			FieldID: field.FieldID, Generation: field.Generation, Variant: field.Variant,
		})
		if err != nil {
			return false
		}
		fields[index] = name
	}
	predicateMatches := actual.predicate == ""
	if expected.ActiveRecordsOnly {
		predicateMatches = actual.predicate == "(deleted_at IS NULL)" ||
			actual.predicate == "deleted_at IS NULL"
	}
	opclassMatches := actual.defaultOperatorClasses
	if expected.OperatorClass == schemav2.OperatorClassGINTrigram {
		opclassMatches = trigramSchema != "" &&
			exactStringValues(actual.operatorClasses,
				string(schemav2.OperatorClassGINTrigram)) &&
			exactStringValues(actual.operatorClassSchemas, trigramSchema)
	}
	return actual.owner == "harpy_owner" &&
		actual.tableName == mapping.Form.PhysicalName &&
		opclassMatches && actual.defaultCollations && actual.defaultOrdering &&
		actual.noIncludedColumns && actual.nullsDistinct && actual.standalone &&
		actual.unique == expected.Unique && actual.method == string(expected.Method) &&
		slices.Equal(actual.fields, fields) && predicateMatches
}

func exactStringValues(values []string, wanted string) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if value != wanted {
			return false
		}
	}
	return true
}
