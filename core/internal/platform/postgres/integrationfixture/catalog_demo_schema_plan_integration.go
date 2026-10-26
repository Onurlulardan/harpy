//go:build integration

package integrationfixture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/google/uuid"
)

const catalogDemoLeaseID = "019c3300-0000-7000-8000-000000000091"
const catalogDemoExecutorID = "019c3300-0000-7000-8000-000000000092"

func (environment *CatalogDemoEnvironment) CompleteCatalogDemoFormsPublication(
	t *testing.T,
	ctx context.Context,
	formID string,
) {
	t.Helper()
	var jobID string
	var planDocument []byte
	err := environment.owner.QueryRowContext(
		ctx, string(catalogDemoFixture("schema_job_load.fixture")), formID,
	).Scan(&jobID, &planDocument)
	if err != nil {
		t.Fatalf("load Catalog demo Forms schema plan: %v", err)
	}
	var plan schemav2.Plan
	if err = json.Unmarshal(planDocument, &plan); err != nil {
		t.Fatalf("decode Catalog demo Forms schema plan: %v", err)
	}
	prepared, err := schemaexecutor.BuildPlanV2(plan)
	if err != nil {
		t.Fatalf("prepare Catalog demo Forms schema plan: %v", err)
	}
	records := catalogDemoStepRecords(plan, prepared)
	for ordinal := range records {
		step := catalogDemoValidatedStep(t, jobID, plan, prepared, records, ordinal)
		if step.Operation() != schemav1.VerifyStorage {
			statement, renderErr := ddl.Render(ddl.RenderContext{
				TenantSchema: CatalogDemoTenantSchema, TrigramSchema: "platform_ext",
			}, step)
			if renderErr != nil {
				t.Fatalf("render Catalog demo schema step %d: %v", ordinal, renderErr)
			}
			if _, err = environment.owner.ExecContext(
				ctx, statement.Text, statement.Arguments...,
			); err != nil {
				t.Fatalf("execute Catalog demo schema step %d: %v", ordinal, err)
			}
		}
		catalogDemoApplyRecordACL(t, ctx, environment, step, ordinal)
	}
	result, err := environment.owner.ExecContext(
		ctx, string(catalogDemoFixture("schema_job_succeed.fixture")), jobID,
	)
	if err != nil {
		t.Fatalf("complete Catalog demo schema job: %v", err)
	}
	if affected, affectErr := result.RowsAffected(); affectErr != nil || affected != 1 {
		t.Fatalf("Catalog demo schema job completion rows=%d err=%v", affected, affectErr)
	}
	environment.completeCatalogDemoPublication(t, ctx, jobID)
}

func catalogDemoStepRecords(
	plan schemav2.Plan,
	prepared schemaexecutor.PreparedPlan,
) []schemaexecutor.StepRecord {
	preparedSteps := prepared.Steps()
	records := make([]schemaexecutor.StepRecord, len(preparedSteps))
	for ordinal, step := range preparedSteps {
		records[ordinal] = schemaexecutor.StepRecord{
			Ordinal: ordinal, Generation: 1, Attempt: 1,
			OperationID:       string(plan.Steps[ordinal].OperationID),
			TransactionClass:  string(plan.Steps[ordinal].TransactionClass),
			ExecutorKind:      string(plan.Steps[ordinal].ExecutorKind),
			IdentifierMapping: step.IdentifierMappingJSON(),
			Arguments:         step.ArgumentsJSON(), StepSHA256: step.SHA256(),
		}
	}
	return records
}

func catalogDemoValidatedStep(
	t *testing.T,
	jobID string,
	plan schemav2.Plan,
	prepared schemaexecutor.PreparedPlan,
	steps []schemaexecutor.StepRecord,
	ordinal int,
) schemaexecutor.ValidatedStep {
	t.Helper()
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	record := schemaexecutor.ClaimRecord{
		Job: schemaexecutor.JobRecord{
			JobID: jobID, TenantID: plan.TenantID, Owner: plan.Owner,
			ResourceKind: string(plan.Resource.Kind), ResourceID: plan.Resource.ID,
			ResourceRevision:    plan.Resource.Revision,
			CompilerVersion:     plan.Compiler.Version,
			PlanContractVersion: plan.ContractVersion,
			CanonicalPlan:       prepared.CanonicalJSON(),
			PlanSHA256:          prepared.SHA256(),
			StepCount:           len(steps), NextStepOrdinal: ordinal,
			Attempt: 1, Generation: 1,
		},
		Steps: steps,
		Lease: schemaexecutor.LeaseRecord{
			JobID: jobID, TenantID: plan.TenantID,
			LeaseID: catalogDemoLeaseID, ExecutorID: catalogDemoExecutorID,
			JobAttempt: 1, StepAttempt: 1, StepOrdinal: ordinal,
			PlanSHA256:    prepared.SHA256(),
			StepSHA256:    prepared.Steps()[ordinal].SHA256(),
			StepStartedAt: now, LeaseExpiresAt: now.Add(time.Minute),
			JobGeneration: 1, StepGeneration: 1,
		},
	}
	_, validated, err := schemaexecutor.AcceptIntegrationClaim(record)
	if err != nil {
		t.Fatalf("validate Catalog demo schema step %d: %v", ordinal, err)
	}
	return validated.CurrentStep()
}

func catalogDemoApplyRecordACL(
	t *testing.T,
	ctx context.Context,
	environment *CatalogDemoEnvironment,
	step schemaexecutor.ValidatedStep,
	ordinal int,
) {
	t.Helper()
	statements, err := ddl.RenderRecordACL(ddl.RenderContext{
		TenantSchema: CatalogDemoTenantSchema, TrigramSchema: "platform_ext",
	}, step)
	if err != nil {
		t.Fatalf("render Catalog demo record ACL step %d: %v", ordinal, err)
	}
	for _, statement := range statements {
		if _, err = environment.owner.ExecContext(
			ctx, statement.Text, statement.Arguments...,
		); err != nil {
			t.Fatalf("execute Catalog demo record ACL step %d: %v", ordinal, err)
		}
	}
}

func (environment *CatalogDemoEnvironment) completeCatalogDemoPublication(
	t *testing.T, ctx context.Context, jobID string,
) {
	t.Helper()
	sum := sha256.Sum256([]byte("forms-publication-apply-result/" + jobID))
	transaction, err := environment.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(
		ctx, string(catalogDemoFixture("runtime_session.fixture")),
	); err != nil {
		t.Fatalf("prepare Catalog demo completion session: %v", err)
	}
	var matched, terminal, activated, remaining int
	var transitions []byte
	err = transaction.QueryRowContext(
		ctx, string(catalogDemoFixture("complete_publication.fixture")),
		jobID, hex.EncodeToString(sum[:]), catalogDemoCompletionID(jobID),
	).Scan(&matched, &terminal, &activated, &remaining, &transitions)
	if err != nil {
		t.Fatalf("complete Catalog demo Forms publication: %v", err)
	}
	if matched != 1 || terminal != 1 || activated != 1 || remaining != 0 ||
		!json.Valid(transitions) {
		t.Fatalf("Catalog demo completion=%d/%d/%d/%d transitions=%s",
			matched, terminal, activated, remaining, transitions)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit Catalog demo Forms publication: %v", err)
	}
}

func catalogDemoCompletionID(jobID string) string {
	digest := sha256.Sum256([]byte("catalog-demo-completion/" + jobID))
	var identifier uuid.UUID
	copy(identifier[:], digest[:16])
	identifier[6] = identifier[6]&0x0f | 0x70
	identifier[8] = identifier[8]&0x3f | 0x80
	return identifier.String()
}
