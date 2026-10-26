package postgres

import (
	"context"
	"slices"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func schemaExecutorVerifyForeignKeys(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	mapping schemav1.IdentifierMapping,
	expected schemav1.RelationArguments,
) error {
	targetTable, err := schemaExecutorFormName(mapping, expected.TargetFormID)
	if err != nil {
		return err
	}
	if err := schemaExecutorRequireTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	); err != nil {
		return err
	}
	if err := schemaExecutorRequireTable(
		ctx, querier, schemaName, targetTable,
	); err != nil {
		return err
	}
	deleteAction := "r"
	if expected.DeletePolicy == schemav1.DeleteCascade {
		deleteAction = "c"
	}
	keys := []struct {
		constraintID string
		sourceField  string
		targetTable  string
	}{
		{expected.SourceConstraintID, "source_id", mapping.Form.PhysicalName},
		{expected.TargetConstraintID, "target_id", targetTable},
	}
	for _, key := range keys {
		name, err := schemaExecutorIdentifierName(mapping.Constraints, key.constraintID)
		if err != nil {
			return err
		}
		actual, found, err := schemaExecutorReadKey(
			ctx, querier, schemaName, tableName, name,
		)
		if err != nil {
			return err
		}
		if !found || actual.owner != "harpy_owner" ||
			actual.kind != "f" || !actual.validated ||
			actual.deleteAction != deleteAction || actual.targetSchema != schemaName ||
			actual.targetTable != key.targetTable || actual.targetOwner != "harpy_owner" ||
			actual.updateAction != "a" || actual.matchType != "s" ||
			actual.deferrable || actual.deferred ||
			!slices.Equal(actual.sourceFields, []string{key.sourceField}) ||
			!slices.Equal(actual.targetFields, []string{"id"}) {
			return schemaExecutorBlocked("schema_relation_foreign_key_mismatch", nil)
		}
	}
	return nil
}

func schemaExecutorVerifyCardinalityKey(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	mapping schemav1.IdentifierMapping,
	expected schemav1.RelationArguments,
) error {
	if expected.CardinalityConstraintID == nil {
		return nil
	}
	name, err := schemaExecutorIdentifierName(
		mapping.Constraints, *expected.CardinalityConstraintID,
	)
	if err != nil {
		return err
	}
	actual, found, err := schemaExecutorReadKey(
		ctx, querier, schemaName, tableName, name,
	)
	if err != nil {
		return err
	}
	if !found || actual.owner != "harpy_owner" ||
		actual.kind != "u" || !actual.validated ||
		actual.deferrable || actual.deferred ||
		!slices.Equal(actual.sourceFields, []string{"source_id"}) {
		return schemaExecutorBlocked("schema_relation_cardinality_mismatch", nil)
	}
	return nil
}
