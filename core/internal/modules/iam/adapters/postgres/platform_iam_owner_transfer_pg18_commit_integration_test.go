//go:build integration

package postgres

import (
	"bytes"
	"context"
	"testing"
)

func exerciseOwnerTransferCommit(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
	attempt ownerTransferAttempt,
) {
	t.Helper()
	prepared := callOwnerTransferPrepare(t, ctx, fixture.runtime, attempt)
	assertTransferResult(t, prepared, attempt, "prepared")
	acknowledged := callOwnerTransferAck(t, ctx, fixture.runtime, attempt, 1)
	assertTransferResult(t, acknowledged, attempt, "ack_verified")
	eventID := "73000000-0000-4000-8000-000000000001"
	eventDigest := digestByte(0x61)
	committed, err := callOwnerTransferCommit(ctx, fixture.runtime, attempt, eventID, eventDigest)
	if err != nil {
		t.Fatalf("commit owner transfer: %v", err)
	}
	assertOwnerTransferCommitResult(t, committed, attempt, false)
	replayed, err := callOwnerTransferCommit(ctx, fixture.runtime, attempt, eventID, eventDigest)
	if err != nil {
		t.Fatalf("replay owner transfer: %v", err)
	}
	assertOwnerTransferCommitResult(t, replayed, attempt, true)
	assertOwnerTransferAuthority(t, ctx, fixture, attempt)
	assertOwnerTransferPrincipals(t, ctx, fixture)
}

func assertOwnerTransferCommitResult(
	t *testing.T, result ownerTransferCommitResult,
	attempt ownerTransferAttempt, replayed bool,
) {
	t.Helper()
	if result.operationID != attempt.operationID ||
		result.predecessorID != attempt.actorID || result.successorID != attempt.successorID ||
		result.predecessorToken != 2 || result.predecessorAuth != 2 ||
		result.successorToken != 2 || result.successorAuth != 2 ||
		result.policyRevision != 2 || result.clock != 2 || result.generation != 2 ||
		result.replayed != replayed {
		t.Fatalf("owner transfer commit result = %+v", result)
	}
}

func assertOwnerTransferAuthority(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
	attempt ownerTransferAttempt,
) {
	t.Helper()
	var binding string
	var policyRevision, clock, generation int64
	var state []byte
	err := fixture.migrator.QueryRowContext(ctx, `
SELECT policy.owner_binding_id,policy.revision,clock.authorization_clock,
       head.generation,head.state_sha256
FROM platform.iam_platform_owner_policies AS policy
JOIN platform.iam_platform_authorization_clocks AS clock USING (environment_id)
JOIN platform.iam_protected_authority_heads AS head USING (environment_id)
WHERE policy.environment_id = $1::uuid`, ownerEnvID).Scan(
		&binding, &policyRevision, &clock, &generation, &state)
	if err != nil {
		t.Fatal(err)
	}
	if binding != attempt.successorBindingID || policyRevision != 2 || clock != 2 ||
		generation != 2 || !bytes.Equal(state, attempt.state) {
		t.Fatalf("authority projection = %s/%d/%d/%d/%x",
			binding, policyRevision, clock, generation, state)
	}
	var predecessorOwner, successorOwner bool
	err = fixture.migrator.QueryRowContext(ctx, `SELECT
    platform.iam_platform_current_owner_v1($1::uuid,$2::uuid),
    platform.iam_platform_current_owner_v1($1::uuid,$3::uuid)`,
		ownerTenantID, ownerUserID, successorID).Scan(&predecessorOwner, &successorOwner)
	if err != nil {
		t.Fatal(err)
	}
	if predecessorOwner || !successorOwner {
		t.Fatalf("owner projection predecessor=%v successor=%v", predecessorOwner, successorOwner)
	}
}

func assertOwnerTransferPrincipals(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
) {
	t.Helper()
	var currentUsers, revokedSessions, auditRows, outboxRows int
	err := fixture.migrator.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM t_owner.iam_users
  WHERE id IN ($1::uuid,$2::uuid) AND token_version = 2
    AND authorization_version = 2 AND revision = 2),
 (SELECT count(*) FROM t_owner.iam_sessions
  WHERE user_id IN ($1::uuid,$2::uuid) AND state = 'revoked'
    AND revision = 2 AND revoked_at IS NOT NULL),
 (SELECT count(*) FROM t_owner.iam_platform_owner_transfer_audit),
 (SELECT count(*) FROM t_owner.iam_platform_owner_transfer_outbox
  WHERE delivery_state = 'pending')`, ownerUserID, successorID).Scan(
		&currentUsers, &revokedSessions, &auditRows, &outboxRows)
	if err != nil {
		t.Fatal(err)
	}
	if currentUsers != 2 || revokedSessions != 2 || auditRows != 1 || outboxRows != 1 {
		t.Fatalf("tenant projection users=%d sessions=%d audit=%d outbox=%d",
			currentUsers, revokedSessions, auditRows, outboxRows)
	}
	if _, err := fixture.migrator.ExecContext(ctx, `UPDATE t_owner.iam_users
SET status = 'disabled', revision = revision + 1, updated_at = clock_timestamp()
WHERE id = $1::uuid`, successorID); err == nil {
		t.Fatal("current Instance Owner lifecycle mutation unexpectedly succeeded")
	}
}
