package schemaexecutor

import (
	"encoding/json"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

var verifyFieldPropertyContract = argumentPropertyContract{
	required: []string{"field_id", "generation", "storage_type", "nullable"},
	optional: []string{"decimal_precision", "decimal_scale"},
}

var verifyRelationPropertyContract = argumentPropertyContract{
	required: []string{
		"relation_id", "target_form_id", "source_constraint_id",
		"target_constraint_id", "cardinality", "delete_policy",
	},
	optional: []string{"cardinality_constraint_id"},
}

var verifyIndexPropertyContract = argumentPropertyContract{required: []string{
	"index_id", "fields", "method", "unique", "active_records_only",
}}

var verifyConstraintPropertyContract = argumentPropertyContract{
	required: []string{"constraint_id", "kind", "fields"},
	optional: []string{"minimum", "maximum"},
}

var fieldReferencePropertyContract = argumentPropertyContract{
	required: []string{"field_id", "generation"},
}

func validateVerifyNestedProperties(arguments schemav1.OperationArguments) error {
	contracts := []struct {
		property string
		contract argumentPropertyContract
		fields   bool
	}{
		{"fields", verifyFieldPropertyContract, false},
		{"relations", verifyRelationPropertyContract, false},
		{"indexes", verifyIndexPropertyContract, true},
		{"constraints", verifyConstraintPropertyContract, true},
	}
	for _, expected := range contracts {
		objects, err := decodeObjectArray(arguments[expected.property])
		if err != nil {
			return fmt.Errorf("verify %s: %w", expected.property, err)
		}
		for index, object := range objects {
			label := fmt.Sprintf("verify %s[%d]", expected.property, index)
			if err := validateObjectProperties(label, object, expected.contract); err != nil {
				return err
			}
			if expected.fields {
				if err := validateFieldReferenceArray(label, object["fields"]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func decodeObjectArray(document json.RawMessage) ([]map[string]json.RawMessage, error) {
	var objects []map[string]json.RawMessage
	if err := json.Unmarshal(document, &objects); err != nil || objects == nil {
		return nil, fmt.Errorf("must be a JSON array of objects")
	}
	return objects, nil
}

func validateFieldReferenceArray(label string, document json.RawMessage) error {
	fields, err := decodeObjectArray(document)
	if err != nil {
		return fmt.Errorf("%s fields: %w", label, err)
	}
	for index, field := range fields {
		fieldLabel := fmt.Sprintf("%s fields[%d]", label, index)
		if err := validateObjectProperties(
			fieldLabel, field, fieldReferencePropertyContract,
		); err != nil {
			return err
		}
	}
	return nil
}
