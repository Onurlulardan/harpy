package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func BuildPlanV2(plan schemav2.Plan) (PreparedPlan, error) {
	if err := validatePlanContractV2(plan); err != nil {
		return PreparedPlan{}, err
	}
	for ordinal, step := range plan.Steps {
		if err := validateResourceOperationV2(plan.Resource.Kind, step.OperationID); err != nil {
			return PreparedPlan{}, fmt.Errorf("step %d: %w", ordinal, err)
		}
		contract, supported := v2OperationContracts[step.OperationID]
		if !supported || step.Ordinal != ordinal || step.ExecutorKind != contract.executor ||
			step.TransactionClass != contract.class {
			return PreparedPlan{}, fmt.Errorf("step %d execution contract differs", ordinal)
		}
		if err := validateIdentifierMappingV2(step.IdentifierMapping, plan.Resource.ID); err != nil {
			return PreparedPlan{}, fmt.Errorf("step %d identifiers: %w", ordinal, err)
		}
		arguments, err := decodeArgumentsV2(step)
		if err != nil {
			return PreparedPlan{}, fmt.Errorf("step %d: %w", ordinal, err)
		}
		if err := validateOperationArgumentsV2(step, arguments); err != nil {
			return PreparedPlan{}, fmt.Errorf("step %d: %w", ordinal, err)
		}
	}
	contractPlan, err := schemav2.PreparePlan(plan)
	if err != nil || len(contractPlan.CanonicalJSON()) > schemav2.MaximumSchemaPlanBytes {
		return PreparedPlan{}, fmt.Errorf("canonical plan is invalid or too large")
	}
	contractSteps := contractPlan.Steps()
	steps := make([]PreparedStep, len(plan.Steps))
	for ordinal, step := range plan.Steps {
		identifiers, identifierErr := marshalCanonical(step.IdentifierMapping)
		arguments, argumentErr := marshalCanonical(step.Arguments)
		if identifierErr != nil || argumentErr != nil {
			return PreparedPlan{}, fmt.Errorf("prepare step %d", ordinal)
		}
		steps[ordinal] = PreparedStep{
			identifierMapping: identifiers, arguments: arguments,
			canonicalPreimage: contractSteps[ordinal].CanonicalPreimage(),
			sha256:            contractSteps[ordinal].SHA256(),
		}
	}
	return PreparedPlan{
		canonical: contractPlan.CanonicalJSON(), sha256: contractPlan.SHA256(), steps: steps,
	}, nil
}
