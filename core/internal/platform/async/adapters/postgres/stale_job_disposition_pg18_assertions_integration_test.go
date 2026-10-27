//go:build integration

package postgres

import (
	"context"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func assertStaleAdoption(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	sourceID uuid.UUID,
	result StaleJobDispositionResult,
	external bool,
) {
	t.Helper()
	resultID, ok := result.ResultJobID()
	if !ok {
		t.Fatal("adoption result omitted replacement job")
	}
	var sourceState, sourceCode string
	var sourceGroup, sourceDomain int64
	var retryOf uuid.UUID
	var resultState, resultWriter, effectClass string
	var resultGroup, resultDomain int64
	var sourceEffect, resultEffect *string
	err := fixture.owner.QueryRowContext(ctx, `
SELECT source.state,source.last_error_code,
       source.authority_group_generation,source.cutover_generation,
       result.retry_of_job_id,result.state,result.writer_origin,
       result.authority_group_generation,result.cutover_generation,
       result.effect_class,source.effect_idempotency_key,
       result.effect_idempotency_key
FROM platform.jobs AS source
JOIN platform.jobs AS result ON result.id=$2
WHERE source.id=$1`, sourceID, resultID).Scan(
		&sourceState, &sourceCode, &sourceGroup, &sourceDomain, &retryOf,
		&resultState, &resultWriter, &resultGroup, &resultDomain, &effectClass,
		&sourceEffect, &resultEffect,
	)
	if err != nil {
		t.Fatal(err)
	}
	if sourceState != "cancelled" || sourceCode != "JOB_AUTHORITY_STALE_ADOPTED" ||
		sourceGroup != 1 || sourceDomain != 1 || retryOf != sourceID ||
		resultState != "queued" || resultWriter != "v2" ||
		resultGroup != 1 || resultDomain != 2 {
		t.Fatalf("adoption rows source=%s/%s/%d/%d result=%s/%s/%d/%d retry=%s",
			sourceState, sourceCode, sourceGroup, sourceDomain, resultState,
			resultWriter, resultGroup, resultDomain, retryOf)
	}
	if external {
		if effectClass != string(jobsv1.EffectExternal) || sourceEffect == nil ||
			resultEffect == nil || *sourceEffect != *resultEffect {
			t.Fatalf("external effect identity = %s/%v/%v", effectClass, sourceEffect, resultEffect)
		}
	} else if effectClass != string(jobsv1.EffectDatabaseTransactional) ||
		sourceEffect != nil || resultEffect != nil {
		t.Fatalf("database effect identity = %s/%v/%v", effectClass, sourceEffect, resultEffect)
	}
	assertStaleDispositionEvidence(t, ctx, fixture, sourceID, result.ActionID(), "adopt")
}

func assertStaleCancellation(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	sourceID uuid.UUID,
	result StaleJobDispositionResult,
) {
	t.Helper()
	if _, ok := result.ResultJobID(); ok {
		t.Fatal("cancellation created a replacement job")
	}
	var state, code string
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT state,last_error_code FROM platform.jobs WHERE id=$1`, sourceID,
	).Scan(&state, &code); err != nil {
		t.Fatal(err)
	}
	if state != "cancelled" || code != "JOB_AUTHORITY_STALE_CANCELLED" {
		t.Fatalf("cancelled stale job = %s/%s", state, code)
	}
	assertStaleDispositionEvidence(t, ctx, fixture, sourceID, result.ActionID(), "cancel")
}

func assertStaleDispositionEvidence(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	sourceID uuid.UUID,
	actionID uuid.UUID,
	decision string,
) {
	t.Helper()
	var payload []byte
	var storedDecision string
	err := fixture.owner.QueryRowContext(ctx, `
SELECT root.payload_canonical,root.redacted_envelope->>'decision'
FROM platform.dead_letters AS root
JOIN platform.dead_letters AS action ON action.root_id=root.id
WHERE root.record_kind='root' AND root.source_kind='job'
  AND root.source_id=$1 AND root.reason_class='stale_generation'
  AND root.authority_group_generation=1 AND root.cutover_generation=2
  AND action.id=$2 AND action.record_kind='repair_action'`,
		sourceID, actionID).Scan(&payload, &storedDecision)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := jobsv1.DecodeStaleJobDisposition(payload)
	if err != nil {
		t.Fatalf("decode stale disposition: %v", err)
	}
	if string(contract.Decision) != decision || storedDecision != decision ||
		contract.SourceJobID != sourceID.String() {
		t.Fatalf("stale disposition evidence = %s/%s/%s",
			contract.Decision, storedDecision, contract.SourceJobID)
	}
}
