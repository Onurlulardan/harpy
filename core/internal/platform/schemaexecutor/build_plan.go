package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func BuildPlan(plan schemav1.Plan) (PreparedPlan, error) {
	if err := validatePlanContract(plan); err != nil {
		return PreparedPlan{}, err
	}
	for ordinal, step := range plan.Steps {
		if err := validateResourceOperation(plan.Resource.Kind, step.OperationID); err != nil {
			return PreparedPlan{}, fmt.Errorf("step %d: %w", ordinal, err)
		}
		contract, supported := operationContracts[step.OperationID]
		if !supported || step.Ordinal != ordinal || step.ExecutorKind != contract.executor ||
			step.TransactionClass != contract.class {
			return PreparedPlan{}, fmt.Errorf("step %d execution contract differs", ordinal)
		}
		if err := validateIdentifierMapping(step.IdentifierMapping, plan.Resource.ID); err != nil {
			return PreparedPlan{}, fmt.Errorf("step %d identifiers: %w", ordinal, err)
		}
		arguments, err := decodeArguments(step)
		if err != nil {
			return PreparedPlan{}, err
		}
		if err := validateOperationArguments(step, arguments); err != nil {
			return PreparedPlan{}, err
		}
	}
	contractPlan, err := schemav1.PreparePlan(plan)
	if err != nil || len(contractPlan.CanonicalJSON()) > schemav1.MaximumSchemaPlanBytes {
		return PreparedPlan{}, fmt.Errorf("canonical plan is invalid or too large")
	}
	contractSteps := contractPlan.Steps()
	steps := make([]PreparedStep, len(plan.Steps))
	for ordinal, step := range plan.Steps {
		prepared, err := buildPreparedStep(contractSteps[ordinal], step)
		if err != nil {
			return PreparedPlan{}, fmt.Errorf("prepare step %d: %w", ordinal, err)
		}
		steps[ordinal] = prepared
	}
	return PreparedPlan{
		canonical: contractPlan.CanonicalJSON(),
		sha256:    contractPlan.SHA256(),
		steps:     steps,
	}, nil
}

func validateResourceOperation(
	resource schemav1.ResourceKind,
	operation schemav1.OperationID,
) error {
	if operation == schemav1.DropFormTable && resource != schemav1.ResourceFormPurge {
		return fmt.Errorf("form table drop requires form-purge resource")
	}
	return nil
}
