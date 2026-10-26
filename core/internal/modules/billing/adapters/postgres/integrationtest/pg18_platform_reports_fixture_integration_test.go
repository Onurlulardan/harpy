//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func preparePlatformReportTenants(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	globalAuthorityID uuid.UUID,
) (billingTenantQueryFixture, billingTenantQueryFixture) {
	t.Helper()
	tenant := prepareBillingTenantFixture(t, ctx, connection, globalAuthorityID)
	subscription := prepareBillingSubscriptionFixture(
		t, ctx, connection, globalAuthorityID, tenant,
	)
	a := billingTenantQueryFixture{
		tenantID: tenantID(), authorityID: tenant.authorityID,
		subscriptionID:    subscription.subscriptionID,
		offeringVersionID: tenant.offeringVersionID,
		counterPeriodID:   tenant.counterPeriodID, gaugePeriodID: tenant.gaugePeriodID,
	}
	yearlyPriceID := billingV7(t)
	if _, err := connection.ExecContext(ctx, `
SELECT platform.billing_add_offering_price_v1(
 $1,$2,$3,'yearly',12018,'EUR',1,$4
)`, globalAuthorityID, tenant.offeringVersionID,
		yearlyPriceID, tenant.now); err != nil {
		t.Fatalf("add yearly report price: %v", err)
	}
	yearly := subscription
	yearly.priceVersionID = yearlyPriceID
	b := prepareBillingTenantQueryB(t, ctx, connection, yearly, tenant.now)
	return a, b
}

func tenantID() uuid.UUID { return uuid.MustParse(billingTenantID) }

func seedPlatformReportAudit(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture platformReportFixture,
	tenant billingTenantQueryFixture,
) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	var beforeRevision, afterRevision int64
	err = transaction.QueryRowContext(ctx, `
UPDATE platform.billing_subscriptions
SET state='past_due',grace_ends_at=$1,updated_at=$2,revision=revision+1
WHERE id=$3 AND tenant_id=$4
RETURNING revision-1,revision
`, now.Add(24*time.Hour), now, tenant.subscriptionID, tenant.tenantID).Scan(
		&beforeRevision, &afterRevision,
	)
	if err != nil {
		t.Fatalf("transition report subscription fixture: %v", err)
	}
	sourceOperationID := billingV7(t)
	if _, err = transaction.ExecContext(ctx, platformReportReceiptSQL,
		billingV7(t), tenant.subscriptionID, sourceOperationID,
		billingDigest("platform-report-request"),
		billingDigest("platform-report-result"), beforeRevision,
		afterRevision, now); err != nil {
		t.Fatalf("seed report operation: %v", err)
	}
	if _, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_subscription_audit (
 id,authority_id,tenant_id,subscription_id,source_operation_id,
 source_kind,kind,actor_id,before_revision,after_revision,occurred_at,reason
) VALUES($1,$2,$3,$4,$5,'actor','billing.subscription.platform_transition',
 $6,$7,$8,$9,'platform report fixture')
`, billingV7(t), tenant.authorityID, tenant.tenantID,
		tenant.subscriptionID, sourceOperationID, fixture.actorID,
		beforeRevision, afterRevision, now); err != nil {
		t.Fatalf("seed report audit: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

const platformReportReceiptSQL = `
INSERT INTO platform.billing_subscription_operations (
 id,authority_id,tenant_id,subscription_id,source_operation_id,
 source_kind,operation_kind,request_sha256,result_sha256,
 expected_revision,result_revision,result_state,completed_at,
 result_provider_account_id,result_offering_version_id,result_price_version_id,
 result_currency,result_period_start,result_period_end,result_trial_ends_at,
 result_grace_ends_at,result_cancel_at_period_end,result_cancel_at,result_updated_at
)
SELECT $1,subscription.authority_id,subscription.tenant_id,subscription.id,$3,
 'actor','billing.subscription.platform_transition',$4,$5,$6,$7,
 subscription.state,$8,subscription.provider_account_id,
 subscription.offering_version_id,subscription.price_version_id,price.currency,
 subscription.period_start,subscription.period_end,subscription.trial_ends_at,
 subscription.grace_ends_at,subscription.cancel_at_period_end,
 subscription.cancel_at,subscription.updated_at
FROM platform.billing_subscriptions AS subscription
JOIN platform.billing_price_versions AS price
  ON price.id=subscription.price_version_id
 AND price.offering_version_id=subscription.offering_version_id
WHERE subscription.id=$2`
