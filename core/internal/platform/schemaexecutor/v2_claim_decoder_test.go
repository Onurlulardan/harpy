package schemaexecutor

import (
	"testing"
	"time"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func TestValidateClaimDispatchesExactV2Contract(t *testing.T) {
	record := testClaimRecordV2(t, testClock())
	claim, err := acceptClaim(record)
	if err != nil {
		t.Fatal(err)
	}
	validated, err := ValidateClaim(claim)
	if err != nil {
		t.Fatal(err)
	}
	step := validated.CurrentStep()
	if step.ContractVersion() != schemav2.PlanContractVersion ||
		step.Operation() != "create_form_table" || validated.StepCount() != 3 {
		t.Fatalf("unexpected V2 validated plan: %#v", step)
	}
	if _, ok := step.V2Mapping(); !ok {
		t.Fatal("V2 mapping was not retained")
	}
}

func TestValidateClaimRejectsUnregisteredContract(t *testing.T) {
	record := testClaimRecordV2(t, testClock())
	record.Job.PlanContractVersion = "forms-schema-plan.v3"
	claim, err := acceptClaim(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateClaim(claim); err == nil {
		t.Fatal("unregistered contract accepted")
	}
}

func testClaimRecordV2(t *testing.T, now time.Time) ClaimRecord {
	t.Helper()
	plan := testTrigramPlanV2(t)
	prepared, err := BuildPlanV2(plan)
	if err != nil {
		t.Fatal(err)
	}
	preparedSteps := prepared.Steps()
	steps := make([]StepRecord, len(preparedSteps))
	for ordinal, step := range preparedSteps {
		steps[ordinal] = StepRecord{
			Ordinal: ordinal, Generation: 2, Attempt: 1,
			OperationID:       string(plan.Steps[ordinal].OperationID),
			TransactionClass:  string(plan.Steps[ordinal].TransactionClass),
			ExecutorKind:      string(plan.Steps[ordinal].ExecutorKind),
			IdentifierMapping: step.IdentifierMappingJSON(),
			Arguments:         step.ArgumentsJSON(), StepSHA256: step.SHA256(),
		}
	}
	return ClaimRecord{
		Job: JobRecord{
			JobID: testJobID, TenantID: testTenantID, Owner: "forms",
			ResourceKind: string(plan.Resource.Kind), ResourceID: testFormID,
			ResourceRevision: 1, CompilerVersion: schemav2.CompilerVersion,
			PlanContractVersion: schemav2.PlanContractVersion,
			CanonicalPlan:       prepared.CanonicalJSON(), PlanSHA256: prepared.SHA256(),
			StepCount: len(steps), NextStepOrdinal: 0, Attempt: 1, Generation: 2,
		},
		Steps: steps,
		Lease: LeaseRecord{
			JobID: testJobID, TenantID: testTenantID, LeaseID: testLeaseID,
			ExecutorID: testExecutorID, JobAttempt: 1, StepAttempt: 1, StepOrdinal: 0,
			PlanSHA256: prepared.SHA256(), StepSHA256: preparedSteps[0].SHA256(),
			StepStartedAt: now, LeaseExpiresAt: now.Add(LeaseDuration),
			JobGeneration: 2, StepGeneration: 2,
		},
	}
}
