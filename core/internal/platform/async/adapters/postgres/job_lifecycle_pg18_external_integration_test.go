//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func assertExternalJobOutcomes(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	successID := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantA, jobsv1.EffectExternal, 8,
	)
	claim := claimOneExternalJob(t, ctx, fixture, successID)
	permit := markExternalEffect(t, ctx, fixture, claim)
	providerReference := "provider/result/42"
	receipt, err := effectsv1.NewSucceededReceipt(
		permit.Request(), uuid.New().String(), "DELIVERED", &providerReference,
		"result/workflow/42", strings.Repeat("b", 64),
		lifecycleReceiptTime(permit.StartedAt()),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.store.CompleteExternalJob(ctx, permit, receipt); err != nil {
		t.Fatal(err)
	}
	state, open := lifecycleState(t, ctx, fixture, successID)
	if state != string(jobsv1.StateSucceeded) || open != 0 {
		t.Fatalf("external success state=%s open=%d", state, open)
	}
	var receiptCount int
	if err := fixture.owner.QueryRowContext(
		ctx, "SELECT count(*) FROM platform.idempotency_receipts WHERE receipt_id=$1",
		receipt.ReceiptID,
	).Scan(&receiptCount); err != nil || receiptCount != 1 {
		t.Fatalf("durable receipt count=%d error=%v", receiptCount, err)
	}

	unknownID := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantA, jobsv1.EffectExternal, 8,
	)
	unknownClaim := claimOneExternalJob(t, ctx, fixture, unknownID)
	unknownPermit := markExternalEffect(t, ctx, fixture, unknownClaim)
	if _, err := fixture.store.MarkExternalJobOutcomeUnknown(
		ctx, unknownPermit,
	); err != nil {
		t.Fatal(err)
	}
	state, open = lifecycleState(t, ctx, fixture, unknownID)
	if state != string(jobsv1.StateDead) || open != 0 {
		t.Fatalf("unknown external state=%s open=%d", state, open)
	}
	assertTerminalJobRoot(
		t, ctx, fixture, unknownID, "outcome_unknown", jobsv1.UnknownOutcomeCode,
	)
	var code string
	if err := fixture.owner.QueryRowContext(
		ctx, "SELECT last_error_code FROM platform.jobs WHERE id=$1", unknownID,
	).Scan(&code); err != nil || code != jobsv1.UnknownOutcomeCode {
		t.Fatalf("unknown outcome code=%q error=%v", code, err)
	}
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, uuid.New(), 10,
		lifecycleCatalog(t, jobsv1.EffectExternal, 8),
	)
	if err != nil || len(claims) != 0 {
		t.Fatalf("unknown outcome was retried: %#v, %v", claims, err)
	}
}

func claimOneExternalJob(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	jobID uuid.UUID,
) JobClaim {
	t.Helper()
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectExternal, 8),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != jobID {
		t.Fatalf("external claim = %#v, %v", claims, err)
	}
	return claims[0]
}

func markExternalEffect(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	claim JobClaim,
) ExternalEffectPermit {
	t.Helper()
	transaction, err := fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	bound := &repository{
		database:  lifecycleBoundPGTx{transaction: transaction},
		authority: lifecycleAuthorityStamp(fixture.authority),
	}
	spec, err := NewExternalEffectRequestSpec(
		uuid.New(), "workflow-provider", "execute",
		"request:"+claim.JobID().String(), uuid.New(),
	)
	if err != nil {
		t.Fatal(err)
	}
	permit, err := bound.MarkExternalEffectStarted(ctx, claim, spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return permit
}
