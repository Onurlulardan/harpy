package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorPrepareConstraintV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	mapping, err := schemaExecutorMappingV2(step)
	if err != nil {
		return false, err
	}
	if err = schemaExecutorRequireTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	); err != nil {
		return false, err
	}
	name, arguments, err := schemaExecutorConstraintIdentityV2(mapping, step)
	if err != nil {
		return false, err
	}
	actual, exists, err := schemaExecutorReadConstraint(
		ctx, querier, schemaName, mapping.Form.PhysicalName, name,
	)
	if err != nil {
		return false, err
	}
	switch step.Operation() {
	case schemav1.AddConstraint:
		if exists {
			return false, schemaExecutorBlocked("schema_object_already_exists", nil)
		}
		return true, nil
	case schemav1.DropConstraint:
		return exists, nil
	case schemav1.ValidateConstraint:
		if !exists {
			return false, schemaExecutorBlocked("schema_constraint_missing", nil)
		}
		if arguments == nil || !schemaExecutorConstraintArgumentsMatchV2(
			actual, mapping, *arguments,
		) {
			return false, schemaExecutorBlocked("schema_constraint_contract_mismatch", nil)
		}
		return !actual.validated, nil
	default:
		return false, fmt.Errorf("V2 constraint precondition is not registered")
	}
}

func schemaExecutorConstraintIdentityV2(
	mapping schemav2.IdentifierMapping,
	step schemaexecutor.ValidatedStep,
) (string, *schemav2.ConstraintArguments, error) {
	if step.Operation() == schemav1.AddConstraint ||
		step.Operation() == schemav1.ValidateConstraint {
		arguments, ok := step.ArgumentValue().(*schemav2.ConstraintArguments)
		if !ok {
			return "", nil, fmt.Errorf("V2 constraint arguments are invalid")
		}
		name, err := schemaExecutorConstraintNameV2(
			mapping.Constraints, arguments.ConstraintID, arguments.Generation,
		)
		return name, arguments, err
	}
	arguments, ok := step.ArgumentValue().(*schemav2.ObjectArguments)
	if !ok {
		return "", nil, fmt.Errorf("V2 constraint object arguments are invalid")
	}
	if arguments.Generation == nil {
		return "", nil, fmt.Errorf("V2 constraint generation is missing")
	}
	name, err := schemaExecutorConstraintNameV2(
		mapping.Constraints, arguments.ObjectID, *arguments.Generation,
	)
	return name, nil, err
}
