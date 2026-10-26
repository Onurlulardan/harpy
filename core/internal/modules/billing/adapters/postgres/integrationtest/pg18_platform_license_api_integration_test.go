//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func assertBillingPlatformLicenseAPI(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	state billingLicenseAcceptedState,
) {
	t.Helper()
	state.snapshot = newBillingLicenseSnapshot(8, "suspended", "terminal-now")
	api := billingLicenseProfileConnection(t, ctx, database, "core-api")
	worker := billingLicenseProfileConnection(t, ctx, database, "core-worker")
	actorID := billingV7(t)
	digest := strings.Repeat("a", 64)
	forceBillingRefreshTargetDue(t, ctx, database, state.fixture)
	first, err := requestPlatformLicenseRefresh(
		ctx, api, actorID, "platform-license-0001", digest,
	)
	if err != nil || first.replayed || first.operationID.Version() != 7 {
		t.Fatalf("platform license request=%+v err=%v", first, err)
	}
	pending, err := readPlatformLicenseRefresh(ctx, api, first.operationID)
	if err != nil || pending.state != "pending" ||
		pending.snapshotSequence.Valid || pending.completedAt.Valid {
		t.Fatalf("pending platform license refresh=%+v err=%v", pending, err)
	}
	replay, err := requestPlatformLicenseRefresh(
		ctx, api, actorID, "platform-license-0001", digest,
	)
	if err != nil || !replay.replayed || replay.operationID != first.operationID ||
		!replay.createdAt.Equal(first.createdAt) {
		t.Fatalf("platform license replay=%+v err=%v", replay, err)
	}
	_, err = requestPlatformLicenseRefresh(
		ctx, api, actorID, "platform-license-0001", strings.Repeat("b", 64),
	)
	if billingPGCode(err) != "23505" {
		t.Fatalf("platform license key conflict code=%q err=%v", billingPGCode(err), err)
	}
	assertPlatformLicenseAccepted(
		t, ctx, database, api, worker, actorID, first, state,
	)
	assertPlatformLicenseOutcomeUnknown(
		t, ctx, database, api, worker, actorID, digest, state,
	)
	assertPlatformLicenseAPIGuards(t, ctx, database, api, actorID)
}

func assertPlatformLicenseAccepted(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	api *sql.Conn,
	worker *sql.Conn,
	actorID uuid.UUID,
	request platformLicenseRefreshRequest,
	state billingLicenseAcceptedState,
) {
	t.Helper()
	target, err := acquireBillingRefreshTarget(ctx, worker)
	if err != nil || !target.ready || target.operationID != request.operationID {
		t.Fatalf("platform license target=%+v err=%v", target, err)
	}
	claim, err := claimBillingLicense(
		ctx, worker, state.fixture, request.operationID, 30,
	)
	if err != nil {
		t.Fatal(err)
	}
	leased, err := readPlatformLicenseRefresh(ctx, api, request.operationID)
	if err != nil || leased.state != "leased" {
		t.Fatalf("leased platform license refresh=%+v err=%v", leased, err)
	}
	_, err = requestPlatformLicenseRefresh(
		ctx, api, actorID, "platform-license-0002", strings.Repeat("c", 64),
	)
	if billingPGCode(err) != "55P03" {
		t.Fatalf("active platform license code=%q err=%v", billingPGCode(err), err)
	}
	assertLicenseAcceptOutcome(
		t, ctx, worker, state.fixture, claim, state.snapshot, "noop",
	)
	accepted, err := readPlatformLicenseRefresh(ctx, api, request.operationID)
	if err != nil || accepted.state != "accepted" ||
		!accepted.snapshotSequence.Valid ||
		accepted.snapshotSequence.Int64 != state.snapshot.sequence ||
		!accepted.snapshotSHA256.Valid ||
		accepted.snapshotSHA256.String != state.snapshot.digest ||
		!accepted.completedAt.Valid {
		t.Fatalf("accepted platform license refresh=%+v err=%v", accepted, err)
	}
	if _, err = database.ExecContext(ctx, `UPDATE
platform.billing_platform_license_refresh_requests SET request_sha256=$1
WHERE operation_id=$2`, strings.Repeat("d", 64), request.operationID); billingPGCode(err) != "55000" {
		t.Fatalf("mutable platform license request code=%q err=%v", billingPGCode(err), err)
	}
	assertPlatformLicenseAdapterRead(t, ctx, api, state)
}
