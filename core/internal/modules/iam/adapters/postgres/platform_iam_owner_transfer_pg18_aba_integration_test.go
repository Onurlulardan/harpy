//go:build integration

package postgres

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func exerciseOwnerTransferCompensation(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
	original ownerTransferAttempt,
) {
	t.Helper()
	prepared := callOwnerTransferPrepare(t, ctx, fixture.runtime, original)
	assertTransferResult(t, prepared, original, "prepared")
	acknowledged := callOwnerTransferAck(t, ctx, fixture.runtime, original, 2)
	assertTransferResult(t, acknowledged, original, "ack_verified")
	compensation := newOwnerCompensationAttempt(original)
	compensationPrepared := callOwnerCompensationPrepare(t, ctx, fixture.runtime, compensation)
	assertOwnerCompensationPrepared(t, compensationPrepared, compensation)
	applied, err := callOwnerCompensationApply(ctx, fixture.runtime, compensation)
	if err != nil {
		t.Fatalf("apply owner compensation: %v", err)
	}
	assertOwnerCompensationApplied(t, applied, compensation, false)
	replayed, err := callOwnerCompensationApply(ctx, fixture.runtime, compensation)
	if err != nil {
		t.Fatalf("replay owner compensation: %v", err)
	}
	assertOwnerCompensationApplied(t, replayed, compensation, true)
	_, err = callOwnerTransferCommit(ctx, fixture.runtime, original,
		"73000000-0000-4000-8000-000000000002", digestByte(0x62))
	assertSerializationFailure(t, err)
	assertOwnerCompensationProjection(t, ctx, fixture, original, compensation)
}

func assertOwnerCompensationPrepared(
	t *testing.T, result ownerTransferResult, attempt ownerCompensationAttempt,
) {
	t.Helper()
	if result.operationID != attempt.operationID || result.registryBindingID != registryID ||
		result.scopeID != ownerEnvID || result.generation != attempt.generation ||
		!bytes.Equal(result.previousHead, attempt.previousHead) ||
		!bytes.Equal(result.state, attempt.state) ||
		!bytes.Equal(result.proposal, attempt.proposal) ||
		result.status != "prepared" || result.replayed {
		t.Fatalf("owner compensation prepare result = %+v", result)
	}
}

func assertOwnerCompensationApplied(
	t *testing.T, result ownerCompensationResult,
	attempt ownerCompensationAttempt, replayed bool,
) {
	t.Helper()
	if result.operationID != attempt.operationID || result.originalID != attempt.originalID ||
		result.generation != attempt.generation || !bytes.Equal(result.state, attempt.state) ||
		!bytes.Equal(result.ack, attempt.ack) || result.replayed != replayed {
		t.Fatalf("owner compensation result = %+v", result)
	}
}

func assertSerializationFailure(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("compensated owner transfer commit unexpectedly succeeded")
	}
	var databaseError *pgconn.PgError
	if !errors.As(err, &databaseError) || databaseError.Code != "40001" {
		t.Fatalf("compensated commit error = %v, want SQLSTATE 40001", err)
	}
}

func assertOwnerCompensationProjection(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
	original ownerTransferAttempt, compensation ownerCompensationAttempt,
) {
	t.Helper()
	var binding, originalStatus, compensationStatus string
	var policyRevision, clock, generation int64
	var state []byte
	err := fixture.migrator.QueryRowContext(ctx, `
SELECT policy.owner_binding_id,policy.revision,clock.authorization_clock,
       head.generation,head.state_sha256,original.status,compensation.status
FROM platform.iam_platform_owner_policies AS policy
JOIN platform.iam_platform_authorization_clocks AS clock USING (environment_id)
JOIN platform.iam_protected_authority_heads AS head USING (environment_id)
JOIN platform.iam_platform_owner_transfer_operations AS original
  ON original.operation_id = $2::uuid
JOIN platform.iam_platform_owner_transfer_compensations AS compensation
  ON compensation.operation_id = $3::uuid
WHERE policy.environment_id = $1::uuid`, ownerEnvID, original.operationID,
		compensation.operationID).Scan(&binding, &policyRevision, &clock,
		&generation, &state, &originalStatus, &compensationStatus)
	if err != nil {
		t.Fatal(err)
	}
	if binding != "63000000-0000-4000-8000-000000000001" ||
		policyRevision != 2 || clock != 2 || generation != compensation.generation ||
		!bytes.Equal(state, original.head) || originalStatus != "compensated" ||
		compensationStatus != "completed" {
		t.Fatalf("compensation projection = %s/%d/%d/%d/%x/%s/%s",
			binding, policyRevision, clock, generation, state,
			originalStatus, compensationStatus)
	}
}
