package postgres

import (
	"context"
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorVerifyRelationKeysV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	mapping schemav2.IdentifierMapping,
	expected schemav2.RelationArguments,
) error {
	primary, found, err := schemaExecutorReadKey(
		ctx, querier, schemaName, tableName, tableName+"_pkey",
	)
	if err != nil {
		return err
	}
	if !found || primary.owner != "harpy_owner" || primary.kind != "p" ||
		!primary.validated || primary.deferrable || primary.deferred ||
		!slices.Equal(primary.sourceFields, []string{"source_id", "target_id"}) {
		return schemaExecutorBlocked("schema_relation_primary_key_mismatch", nil)
	}
	if err = schemaExecutorVerifyForeignKeysV2(
		ctx, querier, schemaName, tableName, mapping, expected,
	); err != nil {
		return err
	}
	return schemaExecutorVerifyRelationSlotsV2(
		ctx, querier, schemaName, tableName, mapping, expected,
	)
}

func schemaExecutorVerifyForeignKeysV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	mapping schemav2.IdentifierMapping,
	expected schemav2.RelationArguments,
) error {
	target, err := schemaExecutorFormNameV2(mapping, expected.TargetFormID)
	if err != nil {
		return err
	}
	deleteAction := "r"
	if expected.DeletePolicy == schemav2.DeleteCascade {
		deleteAction = "c"
	}
	keys := []struct{ id, field, target string }{
		{expected.SourceConstraintID, "source_id", mapping.Form.PhysicalName},
		{expected.TargetConstraintID, "target_id", target},
	}
	for _, key := range keys {
		name, err := schemaExecutorIdentifierNameV2(mapping.Constraints, key.id)
		if err != nil {
			return err
		}
		actual, found, err := schemaExecutorReadKey(
			ctx, querier, schemaName, tableName, name,
		)
		if err != nil {
			return err
		}
		if !found || actual.owner != "harpy_owner" || actual.kind != "f" ||
			!actual.validated || actual.deleteAction != deleteAction ||
			actual.targetSchema != schemaName || actual.targetTable != key.target ||
			actual.targetOwner != "harpy_owner" || actual.updateAction != "a" ||
			actual.matchType != "s" || actual.deferrable || actual.deferred ||
			!slices.Equal(actual.sourceFields, []string{key.field}) ||
			!slices.Equal(actual.targetFields, []string{"id"}) {
			return schemaExecutorBlocked("schema_relation_foreign_key_mismatch", nil)
		}
	}
	return nil
}
