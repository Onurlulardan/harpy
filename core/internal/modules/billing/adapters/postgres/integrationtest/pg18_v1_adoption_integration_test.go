//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
	"github.com/google/uuid"
)

type verifiedEnvelopeFixture struct{}

func (verifiedEnvelopeFixture) Verify(
	_ context.Context,
	operation billingv1adoptionapp.Operation,
	plan billingv1adoptionapp.ScopePlan,
) (billingv1adoptionapp.CutoverEvidence, error) {
	digest := strings.Repeat("ab", 32)
	return billingv1adoptionapp.SealCutoverEvidence(billingv1adoptionapp.CutoverEvidence{
		OperationID: operation.OperationID.String(), RunID: operation.RunID.String(),
		LeaseID: uuid.NewString(), EvidenceGeneration: 1,
		ScopeKind: plan.ScopeKind, ScopeID: plan.ScopeID, PlanSHA256: operation.PlanSHA256,
		LeaseManifestSHA256: digest, BatchSHA256: digest,
		SessionSealReceiptID: uuid.NewString(), SessionSealSHA256: digest,
	})
}

func TestBillingV1AdoptionZeroDiffAndRollbackOnPG18(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := billingTestContext(t)
			defer cancel()
			database := prepareBillingOrigin(t, ctx, origin)
			now := time.Now().UTC().Truncate(time.Microsecond)
			globalAuthority := billingGlobalAuthorityID(t, ctx, database)
			tenantAuthority := registerBillingV1TenantAuthority(t, ctx, database, now)
			switchBillingV1Shadow(t, ctx, database)
			connection := billingAdopterConnection(t, ctx, database)
			store := newBillingV1IntegrationStore(connection)
			snapshot, mapping := billingV1AdoptionFixture(now)
			bundle, err := billingv1adoptionapp.Compile(snapshot, mapping)
			if err != nil || !bundle.Global.Eligible || !bundle.Tenants[0].Eligible {
				t.Fatalf("compile Billing V1 fixture: %v %#v", err, bundle.Global.Issues)
			}
			globalOperation := billingV1Operation(t, bundle.Global, globalAuthority)
			tenantOperation := billingV1Operation(t, bundle.Tenants[0], tenantAuthority)
			verifier := verifiedEnvelopeFixture{}
			for _, value := range []struct {
				operation billingv1adoptionapp.Operation
				plan      billingv1adoptionapp.ScopePlan
			}{{globalOperation, bundle.Global}, {tenantOperation, bundle.Tenants[0]}} {
				if state, registerErr := billingv1adoptionapp.RegisterScope(
					ctx, store, verifier, value.operation, value.plan,
				); registerErr != nil || state != "planned" {
					t.Fatalf("register Billing V1 scope state=%q err=%v", state, registerErr)
				}
			}
			if _, err = billingv1adoptionapp.ApplyScope(ctx, store, globalOperation); err != nil {
				t.Fatal(err)
			}
			assertBillingV1OfferingLifecycle(t, ctx, connection)
			assertBillingV1PromotionLifecycle(t, ctx, connection, now)
			if _, err = billingv1adoptionapp.ApplyScope(ctx, store, tenantOperation); err != nil {
				t.Fatal(err)
			}
			if _, err = billingv1adoptionapp.VerifyScope(ctx, store, tenantOperation); err != nil {
				t.Fatal(err)
			}
			if _, err = billingv1adoptionapp.VerifyScope(ctx, store, globalOperation); err != nil {
				t.Fatal(err)
			}
			assertBillingV1EventArchive(t, ctx, connection)
			assertBillingV1CompatibilityReports(t, ctx, database)
			assertBillingV1Rollback(t, ctx, connection, store, globalOperation, bundle.Global, now)
			assertBillingV1Rollback(t, ctx, connection, store, tenantOperation, bundle.Tenants[0], now)
			assertBillingV1OpeningIdentity(t, ctx, connection)
		})
	}
}

func billingV1Operation(
	t *testing.T, plan billingv1adoptionapp.ScopePlan, authorityID uuid.UUID,
) billingv1adoptionapp.Operation {
	t.Helper()
	_, digest, err := billingv1adoptionapp.EncodeScopePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	return billingv1adoptionapp.Operation{
		OperationID: uuid.New(), RunID: uuid.New(), AuthorityID: authorityID,
		PlanSHA256: digest,
	}
}

func registerBillingV1TenantAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, now time.Time,
) uuid.UUID {
	t.Helper()
	authorityID := billingV7(t)
	connection := billingRuntimeConnection(t, ctx, database)
	if _, err := connection.ExecContext(ctx, `SELECT platform.billing_register_tenant_authority_v1(
$1,$2,1,1,1,$3)`, authorityID, uuid.MustParse(billingTenantID), now); err != nil {
		t.Fatal(err)
	}
	return authorityID
}

func switchBillingV1Shadow(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `
UPDATE platform.cutover_groups SET state='shadow',writer='v1';
UPDATE platform.domain_cutovers SET state='shadow',writer='v1'`); err != nil {
		t.Fatal(err)
	}
}

func billingAdopterConnection(
	t *testing.T, ctx context.Context, database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err == nil {
		_, err = connection.ExecContext(ctx, "SET SESSION AUTHORIZATION harpy_adopter")
	}
	if err == nil {
		_, err = connection.ExecContext(ctx, "SET ROLE harpy_owner")
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	return connection
}
