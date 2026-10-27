package schemaexecutor

import (
	"fmt"
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

var supportedVerifyChecksV2 = []schemav2.VerifyCheck{
	schemav2.VerifyConstraints,
	schemav2.VerifyEvolutionData,
	schemav2.VerifyFields,
	schemav2.VerifyFormTable,
	schemav2.VerifyIndexes,
	schemav2.VerifyRelations,
	schemav2.VerifySystemColumns,
}

func validateVerifyArgumentsV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.VerifyArguments,
) error {
	if len(arguments.Checks) < 1 || len(arguments.Checks) > len(supportedVerifyChecksV2) {
		return fmt.Errorf("verify check count is invalid")
	}
	for index, check := range arguments.Checks {
		if !slices.Contains(supportedVerifyChecksV2, check) {
			return fmt.Errorf("verify check is unknown")
		}
		if index > 0 && arguments.Checks[index-1] >= check {
			return fmt.Errorf("verify checks are not sorted and unique")
		}
	}
	formTable := slices.Contains(arguments.Checks, schemav2.VerifyFormTable)
	systemColumns := slices.Contains(arguments.Checks, schemav2.VerifySystemColumns)
	if formTable != systemColumns {
		return fmt.Errorf("V2 form-table verification requires system columns")
	}
	if err := validateVerifyPresenceV2(arguments); err != nil {
		return err
	}
	if slices.Contains(arguments.Checks, schemav2.VerifyFields) {
		if len(arguments.Fields) != len(mapping.Fields) {
			return fmt.Errorf("verify field contract count differs")
		}
		for index, field := range arguments.Fields {
			mapped := mapping.Fields[index]
			if field.FieldID != mapped.LogicalID || field.Generation != mapped.Generation ||
				field.Variant != mapped.Variant {
				return fmt.Errorf("verify field contract order differs")
			}
			if err := validateVerifyFieldV2(mapping, field); err != nil {
				return err
			}
		}
	}
	if err := validateVerifyRelationsV2(mapping, arguments); err != nil {
		return err
	}
	if err := validateVerifyIndexesV2(mapping, arguments); err != nil {
		return err
	}
	if err := validateVerifyConstraintsV2(mapping, arguments); err != nil {
		return err
	}
	return validateEvolutionDataV2(mapping, arguments)
}

func validateVerifyPresenceV2(arguments schemav2.VerifyArguments) error {
	pairs := []struct {
		check schemav2.VerifyCheck
		count int
	}{
		{schemav2.VerifyFields, len(arguments.Fields)},
		{schemav2.VerifyRelations, len(arguments.Relations)},
		{schemav2.VerifyIndexes, len(arguments.Indexes)},
		{schemav2.VerifyConstraints, len(arguments.Constraints)},
		{schemav2.VerifyEvolutionData, len(arguments.Evolution)},
	}
	for _, pair := range pairs {
		if !slices.Contains(arguments.Checks, pair.check) && pair.count != 0 {
			return fmt.Errorf("verify contract supplied without check")
		}
	}
	return nil
}

func validateVerifyFieldV2(
	mapping schemav2.IdentifierMapping,
	field schemav2.VerifyFieldContract,
) error {
	reference := schemav2.FieldReference{
		FieldID: field.FieldID, Generation: field.Generation, Variant: field.Variant,
	}
	if err := requireFieldV2(mapping, reference); err != nil {
		return err
	}
	if field.Variant != schemav2.VariantValue && field.StorageType != schemav2.StorageText {
		return fmt.Errorf("companion verify field must be text")
	}
	return validateColumnStorageV2(
		field.StorageType, field.DecimalPrecision, field.DecimalScale,
	)
}

func validateVerifyRelationsV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.VerifyArguments,
) error {
	if !slices.Contains(arguments.Checks, schemav2.VerifyRelations) {
		return nil
	}
	if len(arguments.Relations) != len(mapping.Relations) {
		return fmt.Errorf("verify relation contract count differs")
	}
	for index, relation := range arguments.Relations {
		if relation.RelationID != mapping.Relations[index].LogicalID {
			return fmt.Errorf("verify relation contract order differs")
		}
		if err := validateRelationArgumentsV2(mapping, relation); err != nil {
			return err
		}
	}
	return nil
}
