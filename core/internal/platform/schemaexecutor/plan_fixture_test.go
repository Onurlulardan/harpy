package schemaexecutor

import (
	"testing"
	"time"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

const (
	testTenantID   = "00000000-0000-4000-8000-000000000001"
	testFormID     = "00000000-0000-4000-8000-000000000002"
	testJobID      = "00000000-0000-4000-8000-000000000003"
	testExecutorID = "00000000-0000-4000-8000-000000000004"
	testLeaseID    = "00000000-0000-4000-8000-000000000005"
)

func testCreateFormPlan(t *testing.T) schemav1.Plan {
	t.Helper()
	formName, err := deriveObjectIdentifier("frm_", testFormID)
	if err != nil {
		t.Fatal(err)
	}
	mapping := schemav1.IdentifierMapping{
		Form:         schemav1.Identifier{LogicalID: testFormID, PhysicalName: formName},
		RelatedForms: []schemav1.Identifier{}, Fields: []schemav1.FieldIdentifier{},
		Relations: []schemav1.Identifier{}, Indexes: []schemav1.Identifier{},
		Constraints: []schemav1.Identifier{},
	}
	return schemav1.Plan{
		ContractVersion: schemav1.PlanContractVersion,
		TenantID:        testTenantID, Owner: "forms",
		Resource: schemav1.Resource{
			Kind: schemav1.ResourceFormStorage, ID: testFormID, Revision: 1,
		},
		Compiler: schemav1.Compiler{
			Version:         schemav1.InitialCompilerVersion,
			PostgreSQLMajor: schemav1.RequiredPostgreSQLMajor,
		},
		AllowedOperations: []schemav1.OperationID{schemav1.CreateFormTable},
		Steps: []schemav1.Step{{
			Ordinal: 0, OperationID: schemav1.CreateFormTable,
			TransactionClass:  schemav1.Transactional,
			ExecutorKind:      schemav1.MigratorExecutor,
			IdentifierMapping: mapping,
			Arguments:         schemav1.OperationArguments{},
		}},
	}
}

func testClaimRecord(t *testing.T, now time.Time) ClaimRecord {
	t.Helper()
	plan := testCreateFormPlan(t)
	prepared, err := BuildPlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	step := prepared.Steps()[0]
	return ClaimRecord{
		Job: JobRecord{
			JobID: testJobID, TenantID: testTenantID, Owner: "forms",
			ResourceKind: string(plan.Resource.Kind), ResourceID: testFormID,
			ResourceRevision: 1, CompilerVersion: schemav1.InitialCompilerVersion,
			PlanContractVersion: schemav1.PlanContractVersion,
			CanonicalPlan:       prepared.CanonicalJSON(), PlanSHA256: prepared.SHA256(),
			StepCount: 1, NextStepOrdinal: 0, Attempt: 1, Generation: 2,
		},
		Steps: []StepRecord{{
			Ordinal: 0, Generation: 2, Attempt: 1,
			OperationID:       string(schemav1.CreateFormTable),
			TransactionClass:  string(schemav1.Transactional),
			ExecutorKind:      string(schemav1.MigratorExecutor),
			IdentifierMapping: step.IdentifierMappingJSON(),
			Arguments:         step.ArgumentsJSON(), StepSHA256: step.SHA256(),
		}},
		Lease: LeaseRecord{
			JobID: testJobID, TenantID: testTenantID, LeaseID: testLeaseID,
			ExecutorID: testExecutorID, JobAttempt: 1, StepAttempt: 1, StepOrdinal: 0,
			PlanSHA256: prepared.SHA256(), StepSHA256: step.SHA256(),
			StepStartedAt: now, LeaseExpiresAt: now.Add(LeaseDuration),
			JobGeneration: 2, StepGeneration: 2,
		},
	}
}
