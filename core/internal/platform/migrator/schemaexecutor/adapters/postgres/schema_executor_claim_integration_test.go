//go:build integration

package postgres

import (
	"context"
	"sync"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func (environment *schemaExecutorIntegrationEnvironment) claimPlan(
	t *testing.T,
	ctx context.Context,
	executorID string,
	leaseID string,
) (schemaexecutor.ClaimRecord, schemaexecutor.Lease, schemaexecutor.ValidatedPlan) {
	t.Helper()
	record, found, err := environment.store.Claim(ctx, schemaexecutor.ClaimRequest{
		ExecutorID: executorID, LeaseID: leaseID,
	})
	if err != nil || !found {
		t.Fatalf("claim schema job: found=%t err=%v", found, err)
	}
	lease, plan, err := schemaexecutor.AcceptIntegrationClaim(record)
	if err != nil {
		t.Fatalf("accept schema claim: %v", err)
	}
	return record, lease, plan
}

func (environment *schemaExecutorIntegrationEnvironment) startPlan(
	t *testing.T,
	ctx context.Context,
	executorID string,
	leaseID string,
) (schemaexecutor.ClaimRecord, schemaexecutor.Lease, schemaexecutor.ValidatedPlan) {
	t.Helper()
	record, lease, plan := environment.claimPlan(t, ctx, executorID, leaseID)
	if err := environment.store.MarkRunning(ctx, lease); err != nil {
		t.Fatalf("mark schema job running: %v", err)
	}
	return record, lease, plan
}

func runSchemaExecutorClaimRace(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID := schemaExecutorIntegrationUUID(200)
	jobID := schemaExecutorIntegrationUUID(201)
	environment.seedPlan(t, ctx, jobID, schemaExecutorCreatePlan(t, tenantID, formID, 1))
	type result struct {
		record schemaexecutor.ClaimRecord
		found  bool
		err    error
	}
	results := make(chan result, 2)
	start := make(chan struct{})
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func(value int) {
			defer wait.Done()
			<-start
			record, found, err := environment.store.Claim(ctx, schemaexecutor.ClaimRequest{
				ExecutorID: schemaExecutorIntegrationUUID(210 + value),
				LeaseID:    schemaExecutorIntegrationUUID(220 + value),
			})
			results <- result{record: record, found: found, err: err}
		}(index)
	}
	close(start)
	wait.Wait()
	close(results)
	claimed := 0
	var winner schemaexecutor.ClaimRecord
	for outcome := range results {
		if outcome.err != nil {
			t.Fatal(outcome.err)
		}
		if outcome.found {
			claimed++
			winner = outcome.record
			if outcome.record.Job.JobID != jobID {
				t.Fatalf("claimed job = %s", outcome.record.Job.JobID)
			}
		}
	}
	if claimed != 1 {
		t.Fatalf("claim winners = %d, want 1", claimed)
	}
	lease, plan, err := schemaexecutor.AcceptIntegrationClaim(winner)
	if err != nil {
		t.Fatal(err)
	}
	if err := environment.store.MarkRunning(ctx, lease); err != nil {
		t.Fatal(err)
	}
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: plan,
	}); err != nil {
		t.Fatalf("complete race winner: %v", err)
	}
}

func runSchemaExecutorMultiStepAttempts(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID, fieldID := schemaExecutorIntegrationUUID(230), schemaExecutorIntegrationUUID(231)
	environment.seedPlan(t, ctx, schemaExecutorIntegrationUUID(232),
		schemaExecutorCreateAndFieldPlan(t, tenantID, formID, fieldID))
	_, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(233), schemaExecutorIntegrationUUID(234))
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: plan,
	}); err != nil {
		t.Fatalf("apply first schema step: %v", err)
	}
	record, lease, plan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(235), schemaExecutorIntegrationUUID(236))
	if record.Lease.JobAttempt != 2 || record.Lease.StepAttempt != 1 {
		t.Fatalf("second claim attempts job=%d step=%d",
			record.Lease.JobAttempt, record.Lease.StepAttempt)
	}
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: plan,
	}); err != nil {
		t.Fatalf("apply second schema step: %v", err)
	}
}
