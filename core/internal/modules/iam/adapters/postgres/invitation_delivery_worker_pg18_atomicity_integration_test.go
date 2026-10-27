//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

const iamDeliveryClaimSQL = `SELECT artifact_id::text
FROM iam_invitation_delivery_worker_claim_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5)`

const iamDeliveryReceiptSQL = `SELECT code
FROM iam_invitation_delivery_worker_finalize_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,
 'delivered','SMTP_DATA_ACCEPTED',$5,$6,$7)`

func assertIAMInvitationDeliveryAtomicity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	fixture iamInvitationRegistrationFixture,
) {
	t.Helper()
	artifactRef := "iam-delivery://tenant/" + tenant.id + "/artifact/" + fixture.artifactID
	effectID := iamPermissionInvitationID("e1", "9a")
	receiptID := iamPermissionInvitationID("e2", "9a")
	effectKey := "iam.invitation.delivery:" + artifactRef
	messageID := "invite.worker@example.invalid"
	claimedAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	connection, transaction, err := beginIAMPGFirstRuntimeTx(ctx, database, tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	mustWriteInvitationDelivery(
		t, ctx, transaction, fixture, artifactRef, effectID, receiptID,
		effectKey, messageID, claimedAt,
	)
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertIAMInvitationDeliveryCounts(t, ctx, database, tenant.schema, "1|0|0")
	connection, transaction, err = beginIAMPGFirstRuntimeTx(ctx, database, tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	mustWriteInvitationDelivery(
		t, ctx, transaction, fixture, artifactRef, effectID, receiptID,
		effectKey, messageID, claimedAt,
	)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertIAMInvitationDeliveryCounts(t, ctx, database, tenant.schema, "0|0|1")
	if err := rejectIAMPGFirst(
		t, ctx, database, tenant.schema,
		"UPDATE iam_invitation_delivery_worker_receipts SET code='CHANGED'",
	); err == nil {
		t.Fatal("immutable delivery receipt was updated")
	}
}

func mustWriteInvitationDelivery(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	fixture iamInvitationRegistrationFixture,
	artifactRef, effectID, receiptID, effectKey, messageID string,
	completedAt time.Time,
) {
	t.Helper()
	var claimedArtifact string
	if err := transaction.QueryRowContext(
		ctx, iamDeliveryClaimSQL, artifactRef, effectID,
		effectKey, messageID, completedAt,
	).Scan(&claimedArtifact); err != nil {
		t.Fatalf("claim delivery artifact: %s", iamPGErrorDetail(err))
	}
	if claimedArtifact != fixture.artifactID {
		t.Fatalf("claimed artifact=%s want=%s", claimedArtifact, fixture.artifactID)
	}
	var code string
	if err := transaction.QueryRowContext(
		ctx, iamDeliveryReceiptSQL, receiptID, artifactRef, effectID,
		effectKey, messageID, bytes.Repeat([]byte{0x71}, 32), completedAt,
	).Scan(&code); err != nil {
		t.Fatalf("insert delivery receipt: %s", iamPGErrorDetail(err))
	}
	if code != "SMTP_DATA_ACCEPTED" {
		t.Fatalf("delivery receipt code=%s", code)
	}
}

func assertIAMInvitationDeliveryCounts(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	want string,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	var got string
	if err := transaction.QueryRowContext(ctx, `SELECT pg_catalog.concat_ws('|',
 (SELECT pg_catalog.count(*) FROM iam_invitation_delivery_artifacts),
 (SELECT pg_catalog.count(*) FROM iam_invitation_delivery_worker_claims),
	(SELECT pg_catalog.count(*) FROM iam_invitation_delivery_worker_receipts))`).Scan(&got); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("query delivery state: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit delivery state query: %v", err)
	}
	if got != want {
		t.Fatalf("delivery state counts=%s want=%s", got, want)
	}
}
