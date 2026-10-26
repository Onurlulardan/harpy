//go:build integration

package integrationfixture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func (environment *CatalogDemoEnvironment) CompleteCatalogDemoAnalyticsFormsApply(
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
		t.Fatalf("load Analytics Forms schema plan: %v", err)
	}
	var plan schemav2.Plan
	if err = json.Unmarshal(planDocument, &plan); err != nil {
		t.Fatalf("decode Analytics Forms schema plan: %v", err)
	}
	prepared, err := schemaexecutor.BuildPlanV2(plan)
	if err != nil {
		t.Fatalf("prepare Analytics Forms schema plan: %v", err)
	}
	records := catalogDemoStepRecords(plan, prepared)
	for ordinal := range records {
		step := catalogDemoValidatedStep(
			t, jobID, plan, prepared, records, ordinal,
		)
		if step.Operation() != "verify_storage" {
			statement, renderErr := ddl.Render(ddl.RenderContext{
				TenantSchema:  CatalogDemoTenantSchema,
				TrigramSchema: "platform_ext",
			}, step)
			if renderErr != nil {
				t.Fatalf("render Analytics Forms schema step %d: %v", ordinal, renderErr)
			}
			if _, err = environment.owner.ExecContext(
				ctx, statement.Text, statement.Arguments...,
			); err != nil {
				t.Fatalf("execute Analytics Forms schema step %d: %v", ordinal, err)
			}
		}
		catalogDemoApplyRecordACL(t, ctx, environment, step, ordinal)
	}
	result, err := environment.owner.ExecContext(
		ctx, string(catalogDemoFixture("schema_job_succeed.fixture")), jobID,
	)
	if err != nil {
		t.Fatalf("complete Analytics Forms schema job: %v", err)
	}
	if affected, affectErr := result.RowsAffected(); affectErr != nil || affected != 1 {
		t.Fatalf("Analytics Forms schema job completion rows=%d err=%v",
			affected, affectErr)
	}
	environment.completeCatalogDemoAnalyticsFormsApply(t, ctx, jobID)
}

func (environment *CatalogDemoEnvironment) completeCatalogDemoAnalyticsFormsApply(
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
		t.Fatalf("prepare Analytics Forms completion session: %v", err)
	}
	var matched, terminal, activated, remaining int
	var transitions []byte
	err = transaction.QueryRowContext(
		ctx, string(catalogDemoFixture("complete_publication.fixture")),
		jobID, hex.EncodeToString(sum[:]), catalogDemoCompletionID(jobID),
	).Scan(&matched, &terminal, &activated, &remaining, &transitions)
	if err != nil {
		t.Fatalf("record Analytics Forms apply result: %v", err)
	}
	var events []struct {
		State string `json:"state"`
	}
	decodeErr := json.Unmarshal(transitions, &events)
	if matched != 1 || terminal != 1 || activated != 0 || remaining != 0 ||
		decodeErr != nil || len(events) != 1 || events[0].State != "applying" {
		t.Fatalf("Analytics Forms apply=%d/%d/%d/%d transitions=%s err=%v",
			matched, terminal, activated, remaining, transitions, decodeErr)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit Analytics Forms apply result: %v", err)
	}
}
