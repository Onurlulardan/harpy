//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

const billingTenantQueryBID = "72000000-0000-4000-8000-000000000072"

type billingTenantQueryFixture struct {
	tenantID, authorityID, subscriptionID             uuid.UUID
	offeringVersionID, counterPeriodID, gaugePeriodID uuid.UUID
}

func prepareBillingTenantQueries(
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
		tenantID: uuid.MustParse(billingTenantID), authorityID: tenant.authorityID,
		subscriptionID:    subscription.subscriptionID,
		offeringVersionID: tenant.offeringVersionID,
		counterPeriodID:   tenant.counterPeriodID, gaugePeriodID: tenant.gaugePeriodID,
	}
	if _, err := connection.ExecContext(ctx, `
UPDATE platform.billing_usage_totals
SET committed_total=2,reserved_total=1,revision=revision+1,updated_at=$1
WHERE authority_id=$2 AND tenant_id=$3 AND meter_period_id=$4
`, tenant.now, a.authorityID, a.tenantID, a.counterPeriodID); err != nil {
		t.Fatalf("seed tenant A counter total: %v", err)
	}
	if _, err := connection.ExecContext(ctx, `
UPDATE platform.billing_usage_totals
SET gauge_total=4,revision=revision+1,updated_at=$1
WHERE authority_id=$2 AND tenant_id=$3 AND meter_period_id=$4
`, tenant.now, a.authorityID, a.tenantID, a.gaugePeriodID); err != nil {
		t.Fatalf("seed tenant A gauge total: %v", err)
	}
	return a, prepareBillingTenantQueryB(
		t, ctx, connection, subscription, tenant.now,
	)
}

func prepareBillingTenantQueryB(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	a billingSubscriptionFixture,
	now time.Time,
) billingTenantQueryFixture {
	t.Helper()
	fixture := billingTenantQueryFixture{
		tenantID: uuid.MustParse(billingTenantQueryBID), authorityID: billingV7(t),
		subscriptionID: billingV7(t), offeringVersionID: a.offeringVersionID,
		counterPeriodID: billingV7(t), gaugePeriodID: billingV7(t),
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.tenants(id,schema_name,status,deleted_at)
VALUES($1,'t_billing_b','ACTIVE',NULL)
`, fixture.tenantID)
	if err == nil {
		cutover := newBillingCutoverAuthorityFixture(
			"tenant", fixture.tenantID, "billing_forms", 1,
			"v2", "v2", billingAcceptanceA, "billing",
		)
		err = seedBillingCutoverAuthority(ctx, transaction, cutover)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1($1,$2,1,1,$3)
`, fixture.authorityID, fixture.tenantID, now)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_tenant_base_offerings(
 id,authority_id,tenant_id,offering_version_id,effective_start,effective_end,
 source_operation_id,created_at
) VALUES($1,$2,$3,$4,$5,NULL,$6,$5)
`, billingV7(t), fixture.authorityID, fixture.tenantID, fixture.offeringVersionID,
			now.Add(-time.Hour), billingV7(t))
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_subscriptions(
 id,authority_id,tenant_id,provider_account_id,offering_version_id,
 price_version_id,state,period_start,period_end,trial_ends_at,grace_ends_at,
 cancel_at_period_end,cancel_at,created_at,updated_at,revision
) VALUES($1,$2,$3,$4,$5,$6,'active',$7,$8,NULL,NULL,false,NULL,$7,$7,1)
`, fixture.subscriptionID, fixture.authorityID, fixture.tenantID,
			a.providerAccountID, fixture.offeringVersionID, a.priceVersionID,
			now.Add(-time.Hour), now.AddDate(0, 1, 0))
	}
	if err == nil {
		err = seedTenantQueryMeters(ctx, transaction, fixture, now)
	}
	if err != nil {
		t.Fatalf("prepare tenant B query fixture: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func seedTenantQueryMeters(
	ctx context.Context,
	transaction *sql.Tx,
	fixture billingTenantQueryFixture,
	now time.Time,
) error {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	meters := []struct {
		id, definition             uuid.UUID
		key, kind, reset           string
		start, end                 time.Time
		committed, reserved, gauge int64
	}{
		{fixture.counterPeriodID, uuid.MustParse("01900000-0000-7000-8000-000000000005"),
			"forms.submissions", "counter", "calendar_month_utc",
			monthStart, monthStart.AddDate(0, 1, 0), 7, 1, 0},
		{fixture.gaugePeriodID, uuid.MustParse("01900000-0000-7000-8000-000000000003"),
			"documents.storage_bytes", "gauge", "never",
			now.Add(-24 * time.Hour), now.AddDate(100, 0, 0), 0, 0, 11},
	}
	for _, meter := range meters {
		_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.billing_meter_periods(
 id,authority_id,tenant_id,entitlement_definition_id,entitlement_key,meter_kind,
 reset_policy,period_start,period_end,state,closed_at,revision
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'open',NULL,1)
`, meter.id, fixture.authorityID, fixture.tenantID, meter.definition,
			meter.key, meter.kind, meter.reset, meter.start, meter.end)
		if err != nil {
			return err
		}
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_usage_totals(
 authority_id,tenant_id,meter_period_id,entitlement_key,meter_kind,
 committed_total,reserved_total,gauge_total,revision,updated_at
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,1,$9)
`, fixture.authorityID, fixture.tenantID, meter.id, meter.key, meter.kind,
			meter.committed, meter.reserved, meter.gauge, now)
		if err != nil {
			return err
		}
	}
	return nil
}
