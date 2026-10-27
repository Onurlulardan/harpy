//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	platformlicensepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/platformlicense"
	"github.com/google/uuid"
)

func assertPlatformLicenseOutcomeUnknown(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	api *sql.Conn,
	worker *sql.Conn,
	actorID uuid.UUID,
	digest string,
	state billingLicenseAcceptedState,
) {
	t.Helper()
	forceBillingRefreshTargetDue(t, ctx, database, state.fixture)
	request, err := requestPlatformLicenseRefresh(
		ctx, api, actorID, "platform-license-0003", digest,
	)
	if err != nil || request.replayed {
		t.Fatalf("unknown outcome request=%+v err=%v", request, err)
	}
	target, err := acquireBillingRefreshTarget(ctx, worker)
	if err != nil || !target.ready || target.operationID != request.operationID {
		t.Fatalf("unknown outcome target=%+v err=%v", target, err)
	}
	firstClaim, err := claimBillingLicense(
		ctx, worker, state.fixture, request.operationID, 5,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = worker.ExecContext(
		ctx, "SELECT pg_catalog.pg_sleep(5.2)",
	); err != nil {
		t.Fatal(err)
	}
	unknown, err := readPlatformLicenseRefresh(ctx, api, request.operationID)
	if err != nil || unknown.state != "failed" ||
		!unknown.failureCode.Valid ||
		unknown.failureCode.String != "outcome_unknown" {
		t.Fatalf("unknown platform license refresh=%+v err=%v", unknown, err)
	}
	replay, err := requestPlatformLicenseRefresh(
		ctx, api, actorID, "platform-license-0003", digest,
	)
	if err != nil || !replay.replayed || replay.operationID != request.operationID {
		t.Fatalf("unknown outcome replay=%+v err=%v", replay, err)
	}
	forceBillingRefreshTargetDue(t, ctx, database, state.fixture)
	reacquired, err := acquireBillingRefreshTarget(ctx, worker)
	if err != nil || !reacquired.ready ||
		reacquired.operationID != request.operationID {
		t.Fatalf("reacquired platform license target=%+v err=%v", reacquired, err)
	}
	secondClaim, err := claimBillingLicense(
		ctx, worker, state.fixture, request.operationID, 30,
	)
	if err != nil || secondClaim.leaseToken == firstClaim.leaseToken {
		t.Fatalf("reclaimed platform license=%+v err=%v", secondClaim, err)
	}
	assertLicenseAcceptOutcome(
		t, ctx, worker, state.fixture, secondClaim, state.snapshot, "noop",
	)
	recovered, err := readPlatformLicenseRefresh(ctx, api, request.operationID)
	if err != nil || recovered.state != "accepted" ||
		!recovered.snapshotSequence.Valid ||
		recovered.snapshotSequence.Int64 != state.snapshot.sequence {
		t.Fatalf("recovered platform license refresh=%+v err=%v", recovered, err)
	}
}

func assertPlatformLicenseAdapterRead(
	t *testing.T,
	ctx context.Context,
	api *sql.Conn,
	state billingLicenseAcceptedState,
) {
	t.Helper()
	store, err := platformlicensepostgres.NewStore(
		platformlicensepostgres.DatabaseAccess{
			QueryRow: func(
				queryContext context.Context,
				statement string,
				arguments ...any,
			) platformlicensepostgres.Row {
				return api.QueryRowContext(queryContext, statement, arguments...)
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	license, err := store.ReadLicense(ctx, state.fixture.environmentID)
	if err != nil || license.EnvironmentID != state.fixture.environmentID ||
		license.SnapshotSequence != state.snapshot.sequence ||
		license.SnapshotSHA256 != state.snapshot.digest ||
		len(license.Features) == 0 || len(license.Limits) == 0 {
		t.Fatalf("platform license projection=%+v err=%v", license, err)
	}
}

func assertPlatformLicenseAPIGuards(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	api *sql.Conn,
	actorID uuid.UUID,
) {
	t.Helper()
	if _, err := api.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	_, err := api.ExecContext(
		ctx, "SELECT * FROM platform.billing_platform_license_refresh_requests",
	)
	if billingPGCode(err) != "42501" {
		t.Fatalf("direct platform license ledger code=%q err=%v", billingPGCode(err), err)
	}
	if _, err = api.ExecContext(ctx, "RESET ROLE"); err != nil {
		t.Fatal(err)
	}
	if _, err = database.ExecContext(ctx, `UPDATE platform.tenancy_deployment_policy
SET deployment_mode='saas', maximum_customer_tenants=NULL`); err != nil {
		t.Fatal(err)
	}
	_, err = requestPlatformLicenseRefresh(
		ctx, api, actorID, "platform-license-0004", strings.Repeat("e", 64),
	)
	if billingPGCode(err) != "42501" {
		t.Fatalf("SaaS platform license request code=%q err=%v", billingPGCode(err), err)
	}
	if _, err = readPlatformLicenseRefresh(
		ctx, api, billingV7(t),
	); billingPGCode(err) != "42501" {
		t.Fatalf("SaaS platform license read code=%q err=%v", billingPGCode(err), err)
	}
	if _, err = database.ExecContext(ctx, `UPDATE platform.tenancy_deployment_policy
SET deployment_mode='onpremise', maximum_customer_tenants=0`); err != nil {
		t.Fatal(err)
	}
}
