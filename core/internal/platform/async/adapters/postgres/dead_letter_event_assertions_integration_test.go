//go:build integration

package postgres

import (
	"context"
	"testing"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
)

func seedDeadLetterEffectReceipt(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
	tenantID uuid.UUID,
) {
	t.Helper()
	_, err := fixture.runtime.Exec(ctx, `
INSERT INTO platform.idempotency_receipts (
    receipt_id, record_kind, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    request_sha256, retention_class,
    effect_id, effect_kind, effect_revision, effect_key,
    provider, operation, correlation_id, causation_id,
    effect_status, effect_code, effect_result_ref, effect_result_sha256,
    completed_at, retain_until
) VALUES (
    $1, 'effect', $2, 'tenant', $2,
    'operations', 1, 'workflow', 1, 'v2',
    $3, 'external_effect',
    $4, 'event.delivery', 1, 'form-entry/42',
    'rabbitmq', 'publish', 'dead-letter-receipt', $5,
    'succeeded', 'EVENT_DELIVERED', 'broker/confirm/1', $6,
    pg_catalog.clock_timestamp(),
    pg_catalog.clock_timestamp() + interval '90 days'
)`, uuid.New(), tenantID, deadLetterDigest([]byte("request")),
		uuid.New(), uuid.MustParse(asyncCausation), deadLetterDigest([]byte("result")))
	if err != nil {
		t.Fatalf("seed idempotency receipt: %v", err)
	}
}

func readDeadLetterPayload(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
	rootID uuid.UUID,
) []byte {
	t.Helper()
	var payload []byte
	if err := fixture.runtime.QueryRow(
		ctx, `SELECT payload_canonical FROM platform.dead_letters WHERE id=$1`, rootID,
	).Scan(&payload); err != nil {
		t.Fatalf("read dead-letter payload: %v", err)
	}
	return append([]byte(nil), payload...)
}

func assertDeadLetterEventRetryRows(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
	root deadletter.EventRoot,
	actionID uuid.UUID,
) {
	t.Helper()
	var actionPayload []byte
	var reason, ticket string
	if err := fixture.runtime.QueryRow(ctx, `
SELECT payload_canonical, action_reason, action_ticket
FROM platform.dead_letters
WHERE id=$1 AND root_id=$2 AND record_kind='retry_action'`,
		actionID, root.RootID()).Scan(&actionPayload, &reason, &ticket); err != nil {
		t.Fatalf("read event retry action: %v", err)
	}
	if actionPayload != nil || reason != "consumer upgraded" || ticket != "OPS-42" {
		t.Fatal("event retry action did not preserve immutable payload and audit")
	}
	var kind, reference, effectKey, state string
	var attempts int
	if err := fixture.runtime.QueryRow(ctx, `
SELECT kind, payload_ref, effect_idempotency_key, state, attempts
FROM platform.jobs WHERE id=$1`, actionID).Scan(
		&kind, &reference, &effectKey, &state, &attempts,
	); err != nil {
		t.Fatalf("read event retry delivery job: %v", err)
	}
	if kind != "workflow.deadletter_event_retry" ||
		reference != "deadletter:event:"+root.RootID().String() ||
		effectKey != "form-entry/42" || state != "queued" || attempts != 0 {
		t.Fatalf("event retry job = %s %s %s %s %d",
			kind, reference, effectKey, state, attempts)
	}
	var receiptCount int
	if err := fixture.runtime.QueryRow(ctx, `
SELECT count(*) FROM platform.idempotency_receipts
WHERE tenant_id=$1 AND effect_key='form-entry/42'`, *root.TenantID()).Scan(
		&receiptCount,
	); err != nil || receiptCount != 1 {
		t.Fatalf("manual retry deleted receipt: count=%d error=%v", receiptCount, err)
	}
}
