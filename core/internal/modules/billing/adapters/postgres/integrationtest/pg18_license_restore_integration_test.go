//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type billingRestoreRequest struct {
	operationID, receiptID, bindingID uuid.UUID
	receiptDigest                     string
	previousSequence, nextSequence    int64
	previousDigest                    string
	expectedRevision                  int64
	snapshot                          billingLicenseSnapshot
}

type billingRestoreReceipt struct {
	operationID, receiptID, bindingID uuid.UUID
	previousSequence, sequence        int64
	previousDigest, digest            string
	revision                          int64
	replay                            bool
}

func assertBillingLicenseRestore(
	t *testing.T, ctx context.Context, database *sql.DB, connection *sql.Conn,
	state billingLicenseAcceptedState,
) billingLicenseAcceptedState {
	t.Helper()
	restored := newBillingLicenseSnapshot(5, "active", "restored")
	request := billingRestoreRequest{
		operationID: billingV7(t), receiptID: uuid.New(), bindingID: uuid.New(),
		receiptDigest:    strings.Repeat("a", 64),
		previousSequence: state.snapshot.sequence, previousDigest: state.snapshot.digest,
		nextSequence: restored.sequence, expectedRevision: state.revision, snapshot: restored,
	}
	receipt, err := commitBillingRestore(ctx, connection, state.fixture, request)
	if err != nil || receipt.replay || receipt.revision != state.revision+1 {
		t.Fatalf("restore receipt=%+v err=%v", receipt, err)
	}
	assertBillingRestoredReceiptResume(t, ctx, database, state.fixture, request)
	replay, err := commitBillingRestore(ctx, connection, state.fixture, request)
	if err != nil || !replay.replay || replay.sequence != restored.sequence {
		t.Fatalf("restore replay=%+v err=%v", replay, err)
	}
	conflict := request
	conflict.receiptDigest = strings.Repeat("b", 64)
	if _, err = commitBillingRestore(ctx, connection, state.fixture, conflict); billingPGCode(err) != "23505" {
		t.Fatalf("restore conflict code=%q err=%v", billingPGCode(err), err)
	}
	rollback := request
	rollback.operationID, rollback.receiptID = billingV7(t), uuid.New()
	rollback.previousSequence, rollback.nextSequence = 3, 4
	rollback.previousDigest = newBillingLicenseSnapshot(3, "revoked", "terminal").digest
	rollback.snapshot = newBillingLicenseSnapshot(4, "active", "stale-restore")
	if _, err = commitBillingRestore(ctx, connection, state.fixture, rollback); billingPGCode(err) != "40001" {
		t.Fatalf("restore rollback code=%q err=%v", billingPGCode(err), err)
	}
	if got := countBillingLicenseRows(t, ctx, connection, "billing_license_restore_receipts"); got != 1 {
		t.Fatalf("restore receipt count=%d", got)
	}
	state.snapshot, state.revision = restored, receipt.revision
	return state
}

func assertBillingRestoredReceiptResume(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture billingLicenseFixture,
	request billingRestoreRequest,
) {
	t.Helper()
	operator := billingLicenseProfileConnection(
		t, ctx, database, "core-operator-license-reconcile",
	)
	var found bool
	var operationID, receiptID, bindingID, licenseID uuid.UUID
	var organizationID, environmentID uuid.UUID
	var receiptDigest, previousDigest, acceptedDigest string
	var previousSequence, acceptedSequence, revision int64
	var databaseNow sql.NullTime
	err := operator.QueryRowContext(ctx, `SELECT * FROM
platform.billing_restored_license_receipt_read_v1($1,$2)`,
		request.operationID, fixture.environmentID,
	).Scan(
		&found, &operationID, &receiptID, &receiptDigest, &bindingID,
		&licenseID, &organizationID, &environmentID,
		&previousSequence, &previousDigest, &acceptedSequence,
		&acceptedDigest, &revision, &databaseNow,
	)
	if err != nil || !found || operationID != request.operationID ||
		receiptID != request.receiptID || receiptDigest != request.receiptDigest ||
		bindingID != request.bindingID || licenseID != fixture.licenseID ||
		organizationID != fixture.organizationID ||
		environmentID != fixture.environmentID ||
		previousSequence != request.previousSequence ||
		previousDigest != request.previousDigest ||
		acceptedSequence != request.nextSequence ||
		acceptedDigest != request.snapshot.digest ||
		revision != request.expectedRevision+1 ||
		!databaseNow.Valid {
		t.Fatalf("post-commit resume receipt invalid found=%v err=%v", found, err)
	}
	if _, err = operator.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatalf("set restored receipt runtime role: %v", err)
	}
	if _, err = operator.ExecContext(
		ctx,
		`SELECT * FROM platform.billing_license_restore_receipts`,
	); billingPGCode(err) != "42501" {
		t.Fatalf("direct restored receipt read code=%q err=%v", billingPGCode(err), err)
	}
	if _, err = operator.ExecContext(ctx, "RESET ROLE"); err != nil {
		t.Fatalf("reset restored receipt runtime role: %v", err)
	}
}

func commitBillingRestore(
	ctx context.Context, connection *sql.Conn, fixture billingLicenseFixture,
	request billingRestoreRequest,
) (billingRestoreReceipt, error) {
	receipt := billingRestoreReceipt{}
	snapshot := request.snapshot
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.billing_commit_restored_license_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,
$19,$20,$21,$22,$23,$24::jsonb,$25::jsonb,$26,$27,$28::jsonb)`,
		request.operationID, request.receiptID, request.receiptDigest,
		request.bindingID, fixture.licenseID, fixture.organizationID,
		fixture.environmentID, request.previousSequence, request.previousDigest,
		request.nextSequence, snapshot.digest, request.expectedRevision,
		"connected", fixture.issuer, fixture.keyID, snapshot.status,
		snapshot.issuedAt, snapshot.notBefore, snapshot.refreshAfter,
		snapshot.grantUntil, snapshot.graceUntil, snapshot.expiresAt,
		fixture.channel, `{}`, `{}`, snapshot.canonical, snapshot.signed,
		billingLicenseProjection(),
	).Scan(&receipt.operationID, &receipt.receiptID, &receipt.bindingID,
		&receipt.previousSequence, &receipt.previousDigest, &receipt.sequence,
		&receipt.digest, &receipt.revision, &receipt.replay)
	return receipt, err
}
