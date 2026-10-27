package schemaexecutor

import (
	"fmt"
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateVerifyIndexesV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.VerifyArguments,
) error {
	if !slices.Contains(arguments.Checks, schemav2.VerifyIndexes) {
		return nil
	}
	if len(arguments.Indexes) != len(mapping.Indexes) {
		return fmt.Errorf("verify index contract count differs")
	}
	for index, contract := range arguments.Indexes {
		if contract.IndexID != mapping.Indexes[index].LogicalID ||
			contract.Generation != mapping.Indexes[index].Generation {
			return fmt.Errorf("verify index contract order differs")
		}
		if err := validateIndexArgumentsV2(mapping, schemav2.IndexArguments{
			IndexID: contract.IndexID, Generation: contract.Generation,
			Fields: contract.Fields,
			Method: contract.Method, OperatorClass: contract.OperatorClass,
			Unique: contract.Unique, ActiveRecordsOnly: contract.ActiveRecordsOnly,
		}); err != nil {
			return err
		}
	}
	return nil
}

func validateVerifyConstraintsV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.VerifyArguments,
) error {
	if !slices.Contains(arguments.Checks, schemav2.VerifyConstraints) {
		return nil
	}
	relationConstraints := make(map[string]struct{}, len(arguments.Relations)*3)
	for _, relation := range arguments.Relations {
		relationConstraints[relation.SourceConstraintID] = struct{}{}
		relationConstraints[relation.TargetConstraintID] = struct{}{}
		if relation.CardinalityConstraintID != nil {
			relationConstraints[*relation.CardinalityConstraintID] = struct{}{}
		}
	}
	customMappings := make([]schemav2.Identifier, 0, len(mapping.Constraints))
	for _, identifier := range mapping.Constraints {
		if _, relation := relationConstraints[identifier.LogicalID]; !relation {
			customMappings = append(customMappings, identifier)
		}
	}
	if len(arguments.Constraints) != len(customMappings) {
		return fmt.Errorf("verify constraint contract count differs")
	}
	seen := map[string]struct{}{}
	for index, contract := range arguments.Constraints {
		if contract.ConstraintID != customMappings[index].LogicalID ||
			contract.Generation != customMappings[index].Generation {
			return fmt.Errorf("verify constraint contract order differs")
		}
		key := fmt.Sprintf("%s:%d", contract.ConstraintID, contract.Generation)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("verify constraint is duplicated")
		}
		seen[key] = struct{}{}
		if err := validateConstraintArgumentsV2(mapping, schemav2.ConstraintArguments{
			ConstraintID: contract.ConstraintID, Generation: contract.Generation,
			Kind:   contract.Kind,
			Fields: contract.Fields, Minimum: contract.Minimum, Maximum: contract.Maximum,
		}); err != nil {
			return err
		}
	}
	return nil
}
