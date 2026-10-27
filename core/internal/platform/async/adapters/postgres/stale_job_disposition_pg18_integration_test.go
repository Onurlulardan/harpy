//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func TestStaleJobDispositionOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newJobLifecyclePGFixture(t, ctx, origin)
			assertGenericStaleJobDispositions(t, ctx, fixture)
			assertFutureStaleAdoptionSchedule(t, ctx, fixture)
			assertStaleDispositionRejectsNullSchedules(t, ctx, fixture)
			assertStaleDispositionCollisionRollback(t, ctx, fixture)
		})
	}
}

func assertGenericStaleJobDispositions(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	tenantID := addStaleJobTenant(t, ctx, fixture)
	databaseID := insertQueuedLifecycleJob(
		t, ctx, fixture, tenantID.String(), jobsv1.EffectDatabaseTransactional, 8,
	)
	externalID := insertQueuedLifecycleJob(
		t, ctx, fixture, tenantID.String(), jobsv1.EffectExternal, 8,
	)
	cancelID := insertQueuedLifecycleJob(
		t, ctx, fixture, tenantID.String(), jobsv1.EffectDatabaseTransactional, 8,
	)
	setPublisherDomainGeneration(t, ctx, fixture.publisherPGFixture, tenantID, 2)
	if parked, err := fixture.store.ReconcileStaleJobs(ctx, 50); err != nil || parked != 3 {
		t.Fatalf("park disposition candidates = %d, %v", parked, err)
	}
	current, err := fixture.store.ResolveStaleJobAuthority(ctx, databaseID)
	if err != nil {
		t.Fatalf("resolve stale disposition authority: %v", err)
	}

	databaseCommand := staleDispositionCommand(
		t, databaseID, jobsv1.StaleJobAdopt, "adopt database work",
	)
	type dispositionResult struct {
		value StaleJobDispositionResult
		err   error
	}
	results := make(chan dispositionResult, 2)
	for range 2 {
		go func() {
			value, err := fixture.store.DisposeStaleJob(ctx, current, databaseCommand)
			results <- dispositionResult{value: value, err: err}
		}()
	}
	created := 0
	var databaseResult StaleJobDispositionResult
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("concurrent database adoption: %v", result.err)
		}
		if result.value.Created() {
			created++
		}
		databaseResult = result.value
	}
	if created != 1 {
		t.Fatalf("database adoption creators = %d, want 1", created)
	}
	assertStaleAdoption(t, ctx, fixture, databaseID, databaseResult, false)

	externalCommand := staleDispositionCommand(
		t, externalID, jobsv1.StaleJobAdopt, "adopt external work",
	)
	externalResult := disposeStaleConcurrently(
		t, ctx, fixture, current, externalCommand, "external adoption",
	)
	assertStaleAdoption(t, ctx, fixture, externalID, externalResult, true)
	assertSameStaleDispositionResult(t, externalResult,
		disposeStaleAgain(t, ctx, fixture, current, externalCommand), "external")

	cancelCommand := staleDispositionCommand(
		t, cancelID, jobsv1.StaleJobCancel, "cancel obsolete work",
	)
	cancelResult := disposeStaleConcurrently(
		t, ctx, fixture, current, cancelCommand, "stale cancellation",
	)
	assertStaleCancellation(t, ctx, fixture, cancelID, cancelResult)
	assertSameStaleDispositionResult(t, cancelResult,
		disposeStaleAgain(t, ctx, fixture, current, cancelCommand), "cancel")

	resultJobID, _ := databaseResult.ResultJobID()
	claims, err := fixture.store.ClaimJobs(
		ctx, current, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != resultJobID {
		t.Fatalf("claim adopted database job = %#v, %v", claims, err)
	}
	assertStaleDispositionReplay(t, ctx, fixture, current, databaseCommand,
		databaseResult, claims[0], databaseID)
	if parked, err := fixture.store.ReconcileStaleJobs(ctx, 50); err != nil || parked != 0 {
		t.Fatalf("terminal disposition reconciliation = %d, %v", parked, err)
	}
}

func staleDispositionCommand(
	t *testing.T,
	jobID uuid.UUID,
	decision jobsv1.StaleJobDecision,
	reason string,
) StaleJobDispositionCommand {
	t.Helper()
	command, err := NewStaleJobDispositionCommand(
		jobID, decision, reason, "OPS-STALE-42", uuid.New(), uuid.New(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return command
}
