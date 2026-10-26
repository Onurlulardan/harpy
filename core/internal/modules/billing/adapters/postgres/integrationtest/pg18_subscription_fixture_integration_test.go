//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

type billingSubscriptionFixture struct {
	billingTenantFixture
	providerAccountID uuid.UUID
	subscriptionID    uuid.UUID
}

func prepareBillingSubscriptionFixture(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	globalAuthorityID uuid.UUID,
	tenant billingTenantFixture,
) billingSubscriptionFixture {
	t.Helper()
	fixture := billingSubscriptionFixture{
		billingTenantFixture: tenant,
		providerAccountID:    billingV7(t),
		subscriptionID:       billingV7(t),
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_provider_accounts(
 id,authority_id,provider_key,provider_account_key,lifecycle,
 supports_checkout,supports_webhooks,supports_proration,supports_refunds,
 created_at,disabled_at,revision
) VALUES($1,$2,'manual',$3,'active',false,false,false,false,$4,NULL,1)
`, fixture.providerAccountID, globalAuthorityID,
		"fixture-"+fixture.providerAccountID.String(), fixture.now)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_subscriptions(
 id,authority_id,tenant_id,provider_account_id,offering_version_id,
 price_version_id,state,period_start,period_end,trial_ends_at,grace_ends_at,
 cancel_at_period_end,cancel_at,created_at,updated_at,revision
) VALUES($1,$2,$3,$4,$5,$6,'active',$7,$8,NULL,NULL,false,NULL,$7,$7,1)
`, fixture.subscriptionID, fixture.authorityID,
			uuid.MustParse(billingTenantID), fixture.providerAccountID,
			fixture.offeringVersionID, fixture.priceVersionID,
			fixture.now, fixture.now.AddDate(0, 1, 0))
	}
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("prepare Billing subscription fixture: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return fixture
}
