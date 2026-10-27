package schemaexecutor

import (
	"bytes"
	"encoding/json"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

type argumentPropertyContract struct {
	required []string
	optional []string
}

var argumentPropertyContracts = map[schemav1.OperationID]argumentPropertyContract{
	schemav1.CreateFormTable: {}, schemav1.DropFormTable: {},
	schemav1.AddFieldColumn: {required: []string{
		"field_id", "storage_type", "nullable",
	}, optional: []string{"decimal_precision", "decimal_scale"}},
	schemav1.DropFieldColumn: {required: []string{"field_id"}},
	schemav1.AddRelationTable: {required: []string{
		"relation_id", "target_form_id", "source_constraint_id",
		"target_constraint_id", "cardinality", "delete_policy",
	}, optional: []string{"cardinality_constraint_id"}},
	schemav1.DropRelationTable:     {required: []string{"object_id"}},
	schemav1.CreateIndex:           indexPropertyContract(),
	schemav1.CreateIndexConcurrent: indexPropertyContract(),
	schemav1.CleanupInvalidIndex:   indexPropertyContract(),
	schemav1.DropIndex:             {required: []string{"object_id"}},
	schemav1.DropIndexConcurrent:   {required: []string{"object_id"}},
	schemav1.AddConstraint:         constraintPropertyContract(),
	schemav1.ValidateConstraint:    constraintPropertyContract(),
	schemav1.DropConstraint:        {required: []string{"object_id"}},
	schemav1.BackfillFieldBatch: {required: []string{
		"source_field_id", "source_generation", "target_field_id",
		"target_generation", "converter",
	}},
	schemav1.VerifyStorage: {required: []string{
		"checks", "fields", "relations", "indexes", "constraints",
	}},
	schemav1.ActivateGeneration: generationPropertyContract(),
	schemav1.DisableOldWriter:   generationPropertyContract(),
	schemav1.EnableDualWrite:    generationPropertyContract(),
	schemav1.RetainGeneration: {required: []string{
		"form_version_id", "retained_generation", "contract_not_before",
	}},
}

func indexPropertyContract() argumentPropertyContract {
	return argumentPropertyContract{required: []string{
		"index_id", "field_ids", "method", "unique", "active_records_only",
	}}
}

func generationPropertyContract() argumentPropertyContract {
	return argumentPropertyContract{required: []string{
		"form_version_id", "source_generation", "target_generation",
	}}
}

func constraintPropertyContract() argumentPropertyContract {
	return argumentPropertyContract{
		required: []string{"constraint_id", "kind", "field_ids"},
		optional: []string{"minimum", "maximum"},
	}
}

func validateArgumentProperties(
	operation schemav1.OperationID,
	arguments schemav1.OperationArguments,
) error {
	contract, exists := argumentPropertyContracts[operation]
	if !exists {
		return fmt.Errorf("operation is not supported")
	}
	if arguments == nil {
		return fmt.Errorf("%s arguments must be a JSON object", operation)
	}
	if err := validateObjectProperties(
		string(operation), map[string]json.RawMessage(arguments), contract,
	); err != nil {
		return err
	}
	if operation == schemav1.VerifyStorage {
		return validateVerifyNestedProperties(arguments)
	}
	return nil
}

func validateObjectProperties(
	label string,
	object map[string]json.RawMessage,
	contract argumentPropertyContract,
) error {
	if object == nil {
		return fmt.Errorf("%s must be a JSON object", label)
	}
	allowed := make(map[string]struct{}, len(contract.required)+len(contract.optional))
	for _, property := range contract.required {
		allowed[property] = struct{}{}
	}
	for _, property := range contract.optional {
		allowed[property] = struct{}{}
	}
	for _, property := range contract.required {
		value, present := object[property]
		if !present || len(value) == 0 || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("%s required property %s is missing", label, property)
		}
	}
	for property, value := range object {
		if _, present := allowed[property]; !present {
			return fmt.Errorf("%s property %s is not allowed", label, property)
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("%s property %s cannot be null", label, property)
		}
	}
	return nil
}
