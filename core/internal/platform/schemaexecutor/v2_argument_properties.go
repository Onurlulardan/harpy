package schemaexecutor

import (
	"encoding/json"
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

var v2ArgumentPropertyContracts = map[schemav2.OperationID]argumentPropertyContract{
	schemav2.CreateFormTable: {}, schemav2.DropFormTable: {},
	schemav2.AddFieldColumn: {required: []string{
		"field_id", "generation", "variant", "storage_type", "nullable",
	}, optional: []string{"decimal_precision", "decimal_scale"}},
	schemav2.DropFieldColumn: {required: []string{"field_id", "generation", "variant"}},
	schemav2.SetFieldNotNull: {required: []string{"field_id", "generation", "variant"}},
	schemav2.AddRelationTable: {required: []string{
		"relation_id", "target_form_id", "source_constraint_id",
		"target_constraint_id", "cardinality", "delete_policy",
	}, optional: []string{"cardinality_constraint_id"}},
	schemav2.DropRelationTable:     {required: []string{"object_id"}},
	schemav2.CreateIndex:           indexPropertyContractV2(),
	schemav2.CreateIndexConcurrent: indexPropertyContractV2(),
	schemav2.CleanupInvalidIndex:   indexPropertyContractV2(),
	schemav2.DropIndex:             {required: []string{"object_id", "generation"}},
	schemav2.DropIndexConcurrent:   {required: []string{"object_id", "generation"}},
	schemav2.AddConstraint:         constraintPropertyContractV2(),
	schemav2.ValidateConstraint:    constraintPropertyContractV2(),
	schemav2.DropConstraint:        {required: []string{"object_id"}},
	schemav2.BackfillFieldBatch: {required: []string{"source", "target", "converter"},
		optional: []string{"default_literal"}},
	schemav2.VerifyStorage: {required: []string{
		"checks", "fields", "relations", "indexes", "constraints", "evolution",
	}},
	schemav2.ActivateGeneration: generationPropertyContract(),
	schemav2.DisableOldWriter:   generationPropertyContract(),
	schemav2.EnableDualWrite:    generationPropertyContract(),
	schemav2.RetainGeneration: {required: []string{
		"form_version_id", "retained_generation", "minimum_retention_days",
		"zero_reference_days",
	}},
}

func indexPropertyContractV2() argumentPropertyContract {
	return argumentPropertyContract{required: []string{
		"index_id", "generation", "fields", "method", "operator_class", "unique",
		"active_records_only",
	}}
}

func constraintPropertyContractV2() argumentPropertyContract {
	return argumentPropertyContract{
		required: []string{"constraint_id", "generation", "kind", "fields"},
		optional: []string{"minimum", "maximum"},
	}
}

func validateArgumentPropertiesV2(
	operation schemav2.OperationID,
	arguments schemav2.OperationArguments,
) error {
	contract, exists := v2ArgumentPropertyContracts[operation]
	if !exists {
		return fmt.Errorf("operation is not supported")
	}
	if arguments == nil {
		return fmt.Errorf("%s arguments must be a JSON object", operation)
	}
	return validateObjectProperties(
		string(operation), map[string]json.RawMessage(arguments), contract,
	)
}
