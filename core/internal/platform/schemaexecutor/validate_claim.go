package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

const maximumSafeJSONInteger = int64(9007199254740991)

func acceptClaim(record ClaimRecord) (Claim, error) {
	if err := validateLeaseIdentity(record); err != nil {
		return Claim{}, blocked("schema_lease_invalid", err)
	}
	return Claim{record: record, lease: Lease{
		record: record.Lease,
		state: &leaseState{
			jobGeneration:  record.Lease.JobGeneration,
			stepGeneration: record.Lease.StepGeneration,
			expiresAt:      record.Lease.LeaseExpiresAt,
		},
	}}, nil
}

func ValidateClaim(claim Claim) (ValidatedPlan, error) {
	record := claim.record
	canonical, err := canonicalizePlan(record.Job.CanonicalPlan)
	if err != nil {
		return ValidatedPlan{}, blocked("schema_plan_invalid", err)
	}
	digest, err := attestation.DigestSchemaPlan(canonical)
	if err != nil || digest != record.Job.PlanSHA256 {
		return ValidatedPlan{}, blocked("schema_plan_tampered", nil)
	}
	validated, err := decodeClaimPlan(record.Job.PlanContractVersion, canonical, record)
	if err != nil {
		return ValidatedPlan{}, err
	}
	return validated, nil
}

func validateLeaseIdentity(claim ClaimRecord) error {
	lease := claim.Lease
	if lease.JobID != claim.Job.JobID || lease.TenantID != claim.Job.TenantID ||
		lease.PlanSHA256 != claim.Job.PlanSHA256 || lease.JobAttempt != claim.Job.Attempt ||
		lease.JobGeneration != claim.Job.Generation ||
		lease.StepOrdinal != claim.Job.NextStepOrdinal ||
		lease.JobAttempt < 1 || lease.StepAttempt < 1 {
		return fmt.Errorf("lease differs from claimed job")
	}
	for _, value := range []string{lease.JobID, lease.TenantID, lease.LeaseID, lease.ExecutorID} {
		if _, err := canonicalUUID(value); err != nil {
			return fmt.Errorf("lease UUID is invalid")
		}
	}
	if lease.StepStartedAt.IsZero() || lease.LeaseExpiresAt.IsZero() ||
		!lease.StepStartedAt.Before(lease.LeaseExpiresAt) ||
		lease.JobGeneration < 1 || lease.StepGeneration < 1 {
		return fmt.Errorf("lease time range is invalid")
	}
	return nil
}

func validatePlanIdentity(plan schemav1.Plan, job JobRecord) error {
	if err := validatePlanContract(plan); err != nil {
		return err
	}
	if job.PlanContractVersion != plan.ContractVersion || job.Owner != plan.Owner ||
		job.CompilerVersion != plan.Compiler.Version || plan.TenantID != job.TenantID ||
		plan.Resource.ID != job.ResourceID || plan.Resource.Revision != job.ResourceRevision ||
		string(plan.Resource.Kind) != job.ResourceKind {
		return fmt.Errorf("resource identity differs")
	}
	if job.StepCount != len(plan.Steps) ||
		len(job.CanonicalPlan) > schemav1.MaximumSchemaPlanBytes {
		return fmt.Errorf("plan step or size budget differs")
	}
	if job.NextStepOrdinal < 0 || job.NextStepOrdinal >= len(plan.Steps) {
		return fmt.Errorf("current step is outside plan")
	}
	return validateOperationAllowlist(plan)
}

func validatePlanContract(plan schemav1.Plan) error {
	if plan.ContractVersion != schemav1.PlanContractVersion || plan.Owner != "forms" {
		return fmt.Errorf("plan contract or owner differs")
	}
	if plan.Compiler.Version != schemav1.InitialCompilerVersion ||
		plan.Compiler.PostgreSQLMajor != schemav1.RequiredPostgreSQLMajor {
		return fmt.Errorf("compiler is not allowlisted")
	}
	if _, err := canonicalUUID(plan.TenantID); err != nil {
		return fmt.Errorf("tenant identity is invalid")
	}
	if _, err := canonicalUUID(plan.Resource.ID); err != nil ||
		plan.Resource.Revision < 1 || plan.Resource.Revision > maximumSafeJSONInteger {
		return fmt.Errorf("resource identity is invalid")
	}
	if plan.Resource.Kind != schemav1.ResourceFormStorage &&
		plan.Resource.Kind != schemav1.ResourceFormPurge {
		return fmt.Errorf("resource kind is invalid")
	}
	if len(plan.Steps) < 1 || len(plan.Steps) > schemav1.MaximumPlanSteps {
		return fmt.Errorf("plan step budget differs")
	}
	return validateOperationAllowlist(plan)
}
