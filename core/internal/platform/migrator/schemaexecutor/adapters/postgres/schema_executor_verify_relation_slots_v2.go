package postgres

import (
	"context"
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorVerifyRelationSlotsV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	mapping schemav2.IdentifierMapping,
	expected schemav2.RelationArguments,
) error {
	check, found, err := schemaExecutorReadConstraint(
		ctx, querier, schemaName, tableName, tableName+"_edge_slot_check",
	)
	if err != nil {
		return err
	}
	if !found || check.owner != "harpy_owner" || check.kind != "c" ||
		!check.validated || !schemaExecutorSlotCheckMatchesV2(check.definition, expected.Cardinality) {
		return schemaExecutorBlocked("schema_relation_slot_check_mismatch", nil)
	}
	name := tableName + "_slot_key"
	wantedFields := []string{"source_id", "edge_slot"}
	if expected.Cardinality == schemav2.CardinalityOne {
		if expected.CardinalityConstraintID == nil {
			return schemaExecutorBlocked("schema_relation_cardinality_mismatch", nil)
		}
		name, err = schemaExecutorIdentifierNameV2(
			mapping.Constraints, *expected.CardinalityConstraintID,
		)
		if err != nil {
			return err
		}
		wantedFields = []string{"source_id"}
	}
	key, found, err := schemaExecutorReadKey(ctx, querier, schemaName, tableName, name)
	if err != nil {
		return err
	}
	if !found || key.owner != "harpy_owner" || key.kind != "u" ||
		!key.validated || key.deferrable || key.deferred ||
		!slices.Equal(key.sourceFields, wantedFields) {
		return schemaExecutorBlocked("schema_relation_cardinality_mismatch", nil)
	}
	return nil
}

func schemaExecutorSlotCheckMatchesV2(
	definition string,
	cardinality schemav2.Cardinality,
) bool {
	normalized := schemaExecutorNormalizeDefinition(definition)
	if cardinality == schemav2.CardinalityOne {
		return normalized == "checkedge_slot=1"
	}
	return normalized == "checkedge_slotbetween1and100" ||
		normalized == "checkedge_slot>=1andedge_slot<=100"
}
