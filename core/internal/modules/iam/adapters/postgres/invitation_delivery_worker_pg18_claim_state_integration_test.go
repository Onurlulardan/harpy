//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const iamDeliveryReleaseSQL = `SELECT *
FROM iam_invitation_delivery_worker_release_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3)`

const iamDeliveryUnknownSQL = `SELECT *
FROM iam_invitation_delivery_worker_mark_unknown_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4)`

func assertIAMInvitationDeliveryClaimState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	fixture iamInvitationRegistrationFixture,
) {
	t.Helper()
	artifactRef := "iam-delivery://tenant/" + tenant.id + "/artifact/" + fixture.artifactID
	claimedAt := time.Date(2026, 7, 16, 13, 0, 0, 0, time.UTC)
	releaseEffect := iamPermissionInvitationID("e3", "9b")
	releaseKey := "iam.invitation.delivery:release:" + artifactRef
	connection, transaction, err := beginIAMPGFirstRuntimeTx(ctx, database, tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	claimIAMDeliveryForState(
		t, ctx, transaction, artifactRef, releaseEffect,
		releaseKey, "invite.release@example.invalid", claimedAt,
	)
	assertIAMDeliveryBoolean(
		t, ctx, transaction, iamDeliveryReleaseSQL,
		artifactRef, releaseEffect, releaseKey,
	)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	_ = connection.Close()
	assertIAMInvitationDeliveryCounts(t, ctx, database, tenant.schema, "1|0|0")
	unknownEffect := iamPermissionInvitationID("e4", "9b")
	unknownKey := "iam.invitation.delivery:unknown:" + artifactRef
	connection, transaction, err = beginIAMPGFirstRuntimeTx(ctx, database, tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	claimIAMDeliveryForState(
		t, ctx, transaction, artifactRef, unknownEffect,
		unknownKey, "invite.unknown@example.invalid", claimedAt,
	)
	for range 2 {
		assertIAMDeliveryBoolean(
			t, ctx, transaction, iamDeliveryUnknownSQL,
			artifactRef, unknownEffect, unknownKey, claimedAt,
		)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	_ = connection.Close()
	assertIAMInvitationDeliveryCounts(t, ctx, database, tenant.schema, "1|1|0")
}

func claimIAMDeliveryForState(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	artifactRef, effectID, effectKey, messageID string,
	claimedAt time.Time,
) {
	t.Helper()
	var artifactID string
	if err := transaction.QueryRowContext(
		ctx, iamDeliveryClaimSQL, artifactRef, effectID,
		effectKey, messageID, claimedAt,
	).Scan(&artifactID); err != nil {
		t.Fatalf("claim delivery artifact: %s", iamPGErrorDetail(err))
	}
}

func assertIAMDeliveryBoolean(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	query string,
	arguments ...any,
) {
	t.Helper()
	result, err := transaction.ExecContext(ctx, query, arguments...)
	if err != nil {
		t.Fatalf("apply delivery claim state: %s", iamPGErrorDetail(err))
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("read delivery claim state result: %v", err)
	}
	if affected != 1 {
		t.Fatalf("delivery claim state rows=%d want=1", affected)
	}
}
