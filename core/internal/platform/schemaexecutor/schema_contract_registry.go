package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

type claimPlanDecoder func([]byte, ClaimRecord) (ValidatedPlan, error)

var claimPlanDecoders = map[string]claimPlanDecoder{
	schemav1.PlanContractVersion: decodeClaimPlanV1,
	schemav2.PlanContractVersion: decodeClaimPlanV2,
}

func decodeClaimPlan(
	contractVersion string,
	canonical []byte,
	record ClaimRecord,
) (ValidatedPlan, error) {
	decoder, supported := claimPlanDecoders[contractVersion]
	if !supported {
		return ValidatedPlan{}, blocked(
			"schema_plan_invalid", fmt.Errorf("schema contract is not registered"),
		)
	}
	return decoder(canonical, record)
}

func decodeClaimPlanV1(canonical []byte, record ClaimRecord) (ValidatedPlan, error) {
	var plan schemav1.Plan
	if err := decodeStrict(canonical, &plan); err != nil {
		return ValidatedPlan{}, blocked("schema_plan_invalid", err)
	}
	if err := validatePlanIdentity(plan, record.Job); err != nil {
		return ValidatedPlan{}, blocked("schema_plan_stale", err)
	}
	steps, err := validatePlanSteps(plan, record)
	if err != nil {
		return ValidatedPlan{}, blocked("schema_plan_tampered", err)
	}
	return ValidatedPlan{steps: steps, current: record.Job.NextStepOrdinal}, nil
}
