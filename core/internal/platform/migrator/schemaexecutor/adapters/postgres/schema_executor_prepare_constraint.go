package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorPrepareConstraint(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (bool, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return schemaExecutorPrepareConstraintV2(ctx, querier, schemaName, step)
	}
	mapping := step.Mapping()
	if err := schemaExecutorRequireTable(
		ctx, querier, schemaName, mapping.Form.PhysicalName,
	); err != nil {
		return false, err
	}
	var name string
	if step.Operation() == schemav1.AddConstraint ||
		step.Operation() == schemav1.ValidateConstraint {
		arguments := step.ArgumentValue().(*schemav1.ConstraintArguments)
		var err error
		name, err = schemaExecutorIdentifierName(mapping.Constraints, arguments.ConstraintID)
		if err != nil {
			return false, err
		}
	} else {
		arguments := step.ArgumentValue().(*schemav1.ObjectArguments)
		var err error
		name, err = schemaExecutorIdentifierName(mapping.Constraints, arguments.ObjectID)
		if err != nil {
			return false, err
		}
	}
	contract, exists, err := schemaExecutorReadConstraint(
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
		arguments := step.ArgumentValue().(*schemav1.ConstraintArguments)
		if !schemaExecutorConstraintArgumentsMatch(contract, mapping, *arguments) {
			return false, schemaExecutorBlocked("schema_constraint_contract_mismatch", nil)
		}
		return !contract.validated, nil
	default:
		return false, fmt.Errorf("constraint precondition is not registered")
	}
}
