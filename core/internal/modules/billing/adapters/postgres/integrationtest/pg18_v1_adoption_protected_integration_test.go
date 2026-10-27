//go:build integration

package integrationtest

import (
	"errors"
	"testing"
	"time"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestBillingV1AdoptionProtectedCutoverOnPG18(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := billingTestContext(t)
			defer cancel()
			database := prepareBillingOrigin(t, ctx, origin)
			now := time.Now().UTC().Truncate(time.Microsecond)
			registerBillingV1TenantAuthority(t, ctx, database, now)
			switchBillingV1Shadow(t, ctx, database)
			snapshot, mapping := billingV1AdoptionFixture(now)
			bundle, err := billingv1adoptionapp.Compile(snapshot, mapping)
			if err != nil || !bundle.Global.Eligible || len(bundle.Tenants) != 1 ||
				!bundle.Tenants[0].Eligible {
				t.Fatalf("compile protected Billing V1 fixture: %v", err)
			}
			global := newProtectedBillingAuthority(t, "global", bundle.Global,
				"7c000000-0000-4000-8000-000000000001",
				"7c000000-0000-4000-8000-000000000002", "8101")
			tenant := newProtectedBillingAuthority(t, "tenant", bundle.Tenants[0],
				"7c000000-0000-4000-8000-000000000011",
				"7c000000-0000-4000-8000-000000000012", "8201")
			prepareProtectedBillingAuthorities(t, ctx, database, global, tenant)
			assertProtectedBillingAccessDenied(t, ctx, database,
				global.operationID, global.runID)

			globalBatch := newProtectedBillingPlanBatch(
				global, "initial_copy", "0/70", "global-initial")
			if _, _, _, err = executeProtectedBillingBatch(
				ctx, database, global, globalBatch, false, true); err == nil {
				t.Fatal("Billing owner session invoked adopter composite")
			}
			applied, keySHA, stateSHA, err := executeProtectedBillingBatch(
				ctx, database, global, globalBatch, true, false)
			if err != nil || applied != 1 || len(keySHA) != 64 || len(stateSHA) != 64 {
				t.Fatalf("rolled back Billing global apply=%d %q %q err=%v",
					applied, keySHA, stateSHA, err)
			}
			assertProtectedBillingCounts(t, ctx, database, global, 0, 0, 0, 0, 0)
			applied, keySHA, stateSHA, err = executeProtectedBillingBatch(
				ctx, database, global, globalBatch, true, true)
			if err != nil || applied != 1 {
				t.Fatalf("Billing global protected apply=%d err=%v", applied, err)
			}
			assertProtectedBillingRun(t, ctx, database, global, "applied")
			replayed, replayKey, replayState, err := executeProtectedBillingBatch(
				ctx, database, global, globalBatch, true, true)
			if err != nil || replayed != applied || replayKey != keySHA || replayState != stateSHA {
				t.Fatalf("Billing global exact replay=%d %q %q err=%v",
					replayed, replayKey, replayState, err)
			}
			tampered := newProtectedBillingPlanBatch(
				global, "initial_copy", "0/70", "global-tampered")
			if _, _, _, err = executeProtectedBillingBatch(
				ctx, database, global, tampered, true, true); err == nil {
				t.Fatal("Billing global batch tamper replay was accepted")
			}

			isolated := tenant
			isolated.scopeID = billingIsolationTenantID
			isolatedBatch := newProtectedBillingPlanBatch(
				isolated, "complete", "0/80", "tenant-isolated")
			if _, _, _, err = executeProtectedBillingBatch(
				ctx, database, isolated, isolatedBatch, true, true); err == nil {
				t.Fatal("Billing cross-tenant adoption was accepted")
			}
			assertProtectedBillingIsolationEmpty(t, ctx, database)

			ownerConnection, err := database.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer ownerConnection.Close()
			assertBillingV1OfferingLifecycle(t, ctx, ownerConnection)
			assertBillingV1PromotionLifecycle(t, ctx, ownerConnection, now)

			tenantBatch := newProtectedBillingPlanBatch(
				tenant, "complete", "0/80", "tenant-complete")
			applied, keySHA, stateSHA, err = executeProtectedBillingBatch(
				ctx, database, tenant, tenantBatch, true, true)
			if err != nil || applied != 1 {
				var pgError *pgconn.PgError
				_ = errors.As(err, &pgError)
				detail := ""
				if pgError != nil {
					detail = pgError.Detail
				}
				t.Fatalf("Billing tenant protected apply=%d err=%v detail=%s",
					applied, err, detail)
			}
			assertProtectedBillingRun(t, ctx, database, tenant, "rollback_rehearsed")
			replayed, replayKey, replayState, err = executeProtectedBillingBatch(
				ctx, database, tenant, tenantBatch, true, true)
			if err != nil || replayed != applied || replayKey != keySHA || replayState != stateSHA {
				t.Fatalf("Billing tenant exact replay=%d err=%v", replayed, err)
			}

			quiet := newProtectedBillingQuietBatch(global, "global-quiet")
			applied, keySHA, stateSHA, err = executeProtectedBillingBatch(
				ctx, database, global, quiet, true, true)
			if err != nil || applied != 0 {
				t.Fatalf("Billing global quiet completion=%d err=%v", applied, err)
			}
			assertProtectedBillingRun(t, ctx, database, global, "rollback_rehearsed")
			replayed, replayKey, replayState, err = executeProtectedBillingBatch(
				ctx, database, global, quiet, true, true)
			if err != nil || replayed != 0 || replayKey != keySHA || replayState != stateSHA {
				t.Fatalf("Billing global quiet replay=%d err=%v", replayed, err)
			}

			assertProtectedBillingCounts(t, ctx, database, global, 1, 1, 2, 1, 1)
			assertProtectedBillingCounts(t, ctx, database, tenant, 1, 1, 1, 1, 1)
			assertProtectedBillingEvidence(t, ctx, database, global, tenant)
			assertBillingV1EventArchive(t, ctx, ownerConnection)
			assertBillingV1OpeningIdentity(t, ctx, ownerConnection)
			assertBillingV1CompatibilityReports(t, ctx, database)
		})
	}
}
