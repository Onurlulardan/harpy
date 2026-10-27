//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func seedProviderPGSecondTenant(
	t *testing.T,
	ctx context.Context,
	f providerPGFixture,
) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	tenantID := uuid.MustParse("72000000-0000-4000-8000-000000000072")
	authorityID, subscriptionID := billingPGV7(t), billingPGV7(t)
	transaction, err := f.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.tenants(id,schema_name,status,deleted_at)
VALUES($1,'t_provider_two','ACTIVE',NULL)
`, tenantID)
	if err == nil {
		cutover := newBillingCutoverAuthorityFixture(
			"tenant", tenantID, "billing_forms", 1,
			"v2", "v2", billingAcceptanceA, "billing", "forms",
		)
		err = seedBillingCutoverAuthority(ctx, transaction, cutover)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1($1,$2,1,1,$3)
`, authorityID, tenantID, f.now)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_subscriptions(
 id,authority_id,tenant_id,provider_account_id,offering_version_id,
 price_version_id,state,period_start,period_end,cancel_at_period_end,
 created_at,updated_at,revision
) VALUES($1,$2,$3,$4,$5,$6,'active',$7,$8,false,$7,$7,1)
`, subscriptionID, authorityID, tenantID, f.providerID,
			f.offeringVersionID, f.priceVersionID, f.now, f.now.AddDate(0, 1, 0))
	}
	if err != nil {
		t.Fatalf("seed provider second tenant: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return tenantID, authorityID, subscriptionID
}
