package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func decodeClaimPlanV2(canonical []byte, record ClaimRecord) (ValidatedPlan, error) {
	var plan schemav2.Plan
	if err := decodeStrict(canonical, &plan); err != nil {
		return ValidatedPlan{}, blocked("schema_plan_invalid", err)
	}
	if err := validatePlanIdentityV2(plan, record.Job); err != nil {
		return ValidatedPlan{}, blocked("schema_plan_stale", err)
	}
	steps, err := validatePlanStepsV2(plan, record)
	if err != nil {
		return ValidatedPlan{}, blocked("schema_plan_tampered", err)
	}
	return ValidatedPlan{steps: steps, current: record.Job.NextStepOrdinal}, nil
}

func validatePlanIdentityV2(plan schemav2.Plan, job JobRecord) error {
	if err := validatePlanContractV2(plan); err != nil {
		return err
	}
	if job.PlanContractVersion != plan.ContractVersion || job.Owner != plan.Owner ||
		job.CompilerVersion != plan.Compiler.Version || plan.TenantID != job.TenantID ||
		plan.Resource.ID != job.ResourceID || plan.Resource.Revision != job.ResourceRevision ||
		string(plan.Resource.Kind) != job.ResourceKind {
		return fmt.Errorf("resource identity differs")
	}
	if job.StepCount != len(plan.Steps) ||
		len(job.CanonicalPlan) > schemav2.MaximumSchemaPlanBytes {
		return fmt.Errorf("plan step or size budget differs")
	}
	if job.NextStepOrdinal < 0 || job.NextStepOrdinal >= len(plan.Steps) {
		return fmt.Errorf("current step is outside plan")
	}
	return validateOperationAllowlistV2(plan)
}

func validatePlanStepsV2(plan schemav2.Plan, claim ClaimRecord) ([]ValidatedStep, error) {
	if len(claim.Steps) != len(plan.Steps) {
		return nil, fmt.Errorf("persisted step count differs")
	}
	validated := make([]ValidatedStep, len(plan.Steps))
	for ordinal, step := range plan.Steps {
		row := claim.Steps[ordinal]
		contract, supported := v2OperationContracts[step.OperationID]
		if !supported || step.Ordinal != ordinal || row.Ordinal != ordinal ||
			step.ExecutorKind != contract.executor || step.TransactionClass != contract.class {
			return nil, fmt.Errorf("step %d execution contract differs", ordinal)
		}
		if err := validateResourceOperationV2(plan.Resource.Kind, step.OperationID); err != nil {
			return nil, err
		}
		if err := validateIdentifierMappingV2(step.IdentifierMapping, plan.Resource.ID); err != nil {
			return nil, fmt.Errorf("step %d identifiers: %w", ordinal, err)
		}
		arguments, err := decodeArgumentsV2(step)
		if err != nil || validateOperationArgumentsV2(step, arguments) != nil {
			return nil, fmt.Errorf("step %d arguments are invalid", ordinal)
		}
		hash, err := validateStepRowV2(claim.Job.PlanSHA256, step, row)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", ordinal, err)
		}
		if err := validateStepProgress(schemav1.OperationID(step.OperationID), row); err != nil {
			return nil, fmt.Errorf("step %d: %w", ordinal, err)
		}
		validated[ordinal] = ValidatedStep{
			contractVersion: schemav2.PlanContractVersion,
			v2Value:         step, arguments: arguments, hash: hash,
		}
	}
	current := validated[claim.Job.NextStepOrdinal]
	if current.v2Value.ExecutorKind != schemav2.MigratorExecutor ||
		current.hash != claim.Lease.StepSHA256 || current.Ordinal() != claim.Lease.StepOrdinal {
		return nil, fmt.Errorf("claimed step is not the exact migrator step")
	}
	row := claim.Steps[claim.Job.NextStepOrdinal]
	if row.Generation != claim.Lease.StepGeneration || row.Attempt != claim.Lease.StepAttempt {
		return nil, fmt.Errorf("claimed step lease differs")
	}
	return validated, nil
}
