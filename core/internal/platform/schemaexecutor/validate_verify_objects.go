package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateVerifyIndex(
	mapping schemav1.IdentifierMapping,
	index schemav1.VerifyIndexContract,
) error {
	if _, found := findIdentifier(mapping.Indexes, index.IndexID); !found {
		return fmt.Errorf("verify index mapping is missing")
	}
	if len(index.Fields) < 1 || len(index.Fields) > 16 {
		return fmt.Errorf("verify index field count is invalid")
	}
	seen := map[schemav1.FieldReference]struct{}{}
	for _, field := range index.Fields {
		if err := requireFieldGeneration(mapping, field.FieldID, field.Generation); err != nil {
			return err
		}
		if _, exists := seen[field]; exists {
			return fmt.Errorf("verify index field is duplicated")
		}
		seen[field] = struct{}{}
	}
	if index.Method != schemav1.IndexBTree && index.Method != schemav1.IndexGIN {
		return fmt.Errorf("verify index method is invalid")
	}
	if index.Method == schemav1.IndexGIN && (index.Unique || len(index.Fields) != 1) {
		return fmt.Errorf("verify GIN index shape is invalid")
	}
	return nil
}

func validateVerifyConstraints(
	mapping schemav1.IdentifierMapping,
	constraints []schemav1.VerifyConstraintContract,
) error {
	seen := map[string]struct{}{}
	for _, constraint := range constraints {
		if _, found := findIdentifier(mapping.Constraints, constraint.ConstraintID); !found {
			return fmt.Errorf("verify constraint mapping is missing")
		}
		if _, exists := seen[constraint.ConstraintID]; exists {
			return fmt.Errorf("verify constraint is duplicated")
		}
		seen[constraint.ConstraintID] = struct{}{}
		if err := validateVerifyConstraint(mapping, constraint); err != nil {
			return err
		}
	}
	return nil
}

func validateVerifyConstraint(
	mapping schemav1.IdentifierMapping,
	constraint schemav1.VerifyConstraintContract,
) error {
	fields := make([]string, len(constraint.Fields))
	for index, field := range constraint.Fields {
		if err := requireFieldGeneration(mapping, field.FieldID, field.Generation); err != nil {
			return err
		}
		fields[index] = field.FieldID
	}
	return validateConstraintArguments(mapping, schemav1.ConstraintArguments{
		ConstraintID: constraint.ConstraintID, Kind: constraint.Kind,
		FieldIDs: fields, Minimum: constraint.Minimum, Maximum: constraint.Maximum,
	})
}
