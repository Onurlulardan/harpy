//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
)

func TestDeadLetterJobRetryAndExpiredPayloadAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newDeadLetterPGFixture(t, ctx, origin)
			assertDeadLetterJobRetry(t, ctx, fixture)
			assertExpiredDeadLetterRetryDenied(t, ctx, fixture)
		})
	}
}

func assertDeadLetterJobRetry(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
) {
	t.Helper()
	tenantID := uuid.MustParse(asyncTenantA)
	sourceJobID, rootID := seedDeadJobRoot(t, ctx, fixture, tenantID)
	resultJobID := uuid.New()
	operator, _ := deadletter.NewOperator(
		deadLetterPermissionFixture{allowed: map[uuid.UUID]struct{}{tenantID: {}}},
		fixture.repository, deadLetterFixedID{value: resultJobID},
	)
	result, err := operator.Retry(ctx, deadletter.RetryRequest{
		ActorID: uuid.New(), TenantID: tenantID, RootID: rootID,
		Reason: "provider issue repaired", Ticket: "OPS-84", CausationID: uuid.New(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceKind != deadletter.SourceJob || result.SourceID != sourceJobID ||
		result.ResultID != resultJobID || result.DeliveryJobID != nil {
		t.Fatalf("job retry result = %#v", result)
	}
	var sourceAttempts, retryAttempts int
	var retryOf uuid.UUID
	var sourceEffect, retryEffect, retryPayload string
	if err := fixture.runtime.QueryRow(ctx, `
SELECT source.attempts, retry.attempts, retry.retry_of_job_id,
       source.effect_idempotency_key, retry.effect_idempotency_key,
       retry.payload_ref
FROM platform.jobs AS source
JOIN platform.jobs AS retry ON retry.id=$2
WHERE source.id=$1`, sourceJobID, resultJobID).Scan(
		&sourceAttempts, &retryAttempts, &retryOf,
		&sourceEffect, &retryEffect, &retryPayload,
	); err != nil {
		t.Fatalf("read job retry pair: %v", err)
	}
	if sourceAttempts != 8 || retryAttempts != 0 || retryOf != sourceJobID ||
		sourceEffect != retryEffect || retryPayload != "jobs/source/42" {
		t.Fatal("job retry reset source attempts or changed payload/effect identity")
	}
}

func assertExpiredDeadLetterRetryDenied(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
) {
	t.Helper()
	tenantID := uuid.MustParse(asyncTenantA)
	rootID := seedExpiredEventRoot(t, ctx, fixture, tenantID)
	operator, _ := deadletter.NewOperator(
		deadLetterPermissionFixture{allowed: map[uuid.UUID]struct{}{tenantID: {}}},
		fixture.repository, deadLetterFixedID{value: uuid.New()},
	)
	_, err := operator.Retry(ctx, deadletter.RetryRequest{
		ActorID: uuid.New(), TenantID: tenantID, RootID: rootID,
		Reason: "late retry", Ticket: "OPS-EXPIRED", CausationID: uuid.New(),
	})
	if !errors.Is(err, ErrDeadLetterPayloadExpired) {
		t.Fatalf("expired payload retry error = %v", err)
	}
	var actions int
	if err := fixture.runtime.QueryRow(ctx, `
SELECT count(*) FROM platform.dead_letters
WHERE root_id=$1 AND record_kind='retry_action'`, rootID).Scan(&actions); err != nil || actions != 0 {
		t.Fatalf("expired retry wrote action: count=%d error=%v", actions, err)
	}
}
