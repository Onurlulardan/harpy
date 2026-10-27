package postgres

import (
	"context"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func schemaExecutorVerifyConstraintsV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav2.IdentifierMapping,
	expected []schemav2.VerifyConstraintContract,
) error {
	for _, constraint := range expected {
		name, err := schemaExecutorConstraintNameV2(
			mapping.Constraints, constraint.ConstraintID, constraint.Generation,
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
			!schemaExecutorVerifyConstraintMatchV2(actual, mapping, constraint) {
			return schemaExecutorBlocked("schema_constraint_invalid", nil)
		}
	}
	return nil
}
