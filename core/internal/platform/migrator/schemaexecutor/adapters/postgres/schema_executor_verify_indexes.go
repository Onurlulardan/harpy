package postgres

import (
	"context"
	"slices"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func schemaExecutorVerifyIndexes(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav1.IdentifierMapping,
	expected []schemav1.VerifyIndexContract,
) error {
	for _, index := range expected {
		name, err := schemaExecutorIdentifierName(mapping.Indexes, index.IndexID)
		if err != nil {
			return err
		}
		contract, exists, err := schemaExecutorReadIndex(
			ctx, querier, schemaName, name,
		)
		if err != nil {
			return err
		}
		if !exists || !schemaExecutorVerifyIndexMatch(contract, mapping, index) {
			return schemaExecutorBlocked("schema_index_invalid", nil)
		}
	}
	return nil
}

func schemaExecutorVerifyIndexMatch(
	contract schemaExecutorIndexContract,
	mapping schemav1.IdentifierMapping,
	expected schemav1.VerifyIndexContract,
) bool {
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
	predicate := contract.predicate == ""
	if expected.ActiveRecordsOnly {
		predicate = contract.predicate == "(deleted_at IS NULL)" ||
			contract.predicate == "deleted_at IS NULL"
	}
	return contract.owner == "harpy_owner" && contract.valid && contract.ready &&
		contract.tableName == mapping.Form.PhysicalName &&
		contract.defaultOperatorClasses && contract.defaultCollations &&
		contract.defaultOrdering && contract.noIncludedColumns && contract.nullsDistinct &&
		contract.standalone &&
		contract.unique == expected.Unique && contract.method == string(expected.Method) &&
		slices.Equal(contract.fields, fields) && predicate
}

func schemaExecutorVerifyConstraints(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav1.IdentifierMapping,
	expected []schemav1.VerifyConstraintContract,
) error {
	for _, constraint := range expected {
		name, err := schemaExecutorIdentifierName(
			mapping.Constraints, constraint.ConstraintID,
		)
		if err != nil {
			return err
		}
		actual, exists, err := schemaExecutorReadConstraint(
			ctx, querier, schemaName, mapping.Form.PhysicalName, name,
		)
		if err != nil {
			return err
		}
		if !exists || !actual.validated ||
			!schemaExecutorConstraintMatches(actual, mapping, constraint) {
			return schemaExecutorBlocked("schema_constraint_invalid", nil)
		}
	}
	return nil
}
