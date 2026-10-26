package schemaexecutor

import (
	"fmt"
	"slices"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateVerifyContracts(
	mapping schemav1.IdentifierMapping,
	arguments schemav1.VerifyArguments,
) error {
	if err := validateVerifyPresence(arguments); err != nil {
		return err
	}
	if slices.Contains(arguments.Checks, schemav1.VerifyFields) {
		if len(arguments.Fields) != len(mapping.Fields) {
			return fmt.Errorf("verify field contract count differs")
		}
		for index, field := range arguments.Fields {
			mapped := mapping.Fields[index]
			if field.FieldID != mapped.LogicalID || field.Generation != mapped.Generation {
				return fmt.Errorf("verify field contract order differs")
			}
			if err := validateVerifyField(mapping, field); err != nil {
				return err
			}
		}
	}
	if slices.Contains(arguments.Checks, schemav1.VerifyRelations) {
		if len(arguments.Relations) != len(mapping.Relations) {
			return fmt.Errorf("verify relation contract count differs")
		}
		for index, relation := range arguments.Relations {
			if relation.RelationID != mapping.Relations[index].LogicalID {
				return fmt.Errorf("verify relation contract order differs")
			}
			if err := validateRelationArguments(mapping, relation); err != nil {
				return err
			}
		}
	}
	if slices.Contains(arguments.Checks, schemav1.VerifyIndexes) {
		if len(arguments.Indexes) != len(mapping.Indexes) {
			return fmt.Errorf("verify index contract count differs")
		}
		for index, expected := range arguments.Indexes {
			if expected.IndexID != mapping.Indexes[index].LogicalID {
				return fmt.Errorf("verify index contract order differs")
			}
			if err := validateVerifyIndex(mapping, expected); err != nil {
				return err
			}
		}
	}
	if slices.Contains(arguments.Checks, schemav1.VerifyConstraints) {
		if len(arguments.Constraints) != len(mapping.Constraints) {
			return fmt.Errorf("verify constraint contract count differs")
		}
		for index, expected := range arguments.Constraints {
			if expected.ConstraintID != mapping.Constraints[index].LogicalID {
				return fmt.Errorf("verify constraint contract order differs")
			}
		}
		return validateVerifyConstraints(mapping, arguments.Constraints)
	}
	return nil
}

func validateVerifyPresence(arguments schemav1.VerifyArguments) error {
	pairs := []struct {
		check schemav1.VerifyCheck
		count int
	}{
		{schemav1.VerifyFields, len(arguments.Fields)},
		{schemav1.VerifyRelations, len(arguments.Relations)},
		{schemav1.VerifyIndexes, len(arguments.Indexes)},
		{schemav1.VerifyConstraints, len(arguments.Constraints)},
	}
	for _, pair := range pairs {
		if !slices.Contains(arguments.Checks, pair.check) && pair.count != 0 {
			return fmt.Errorf("verify contract supplied without check")
		}
	}
	return nil
}

func validateVerifyField(
	mapping schemav1.IdentifierMapping,
	field schemav1.VerifyFieldContract,
) error {
	if err := requireFieldGeneration(mapping, field.FieldID, field.Generation); err != nil {
		return err
	}
	return validateColumnStorage(
		field.StorageType, field.DecimalPrecision, field.DecimalScale,
	)
}
