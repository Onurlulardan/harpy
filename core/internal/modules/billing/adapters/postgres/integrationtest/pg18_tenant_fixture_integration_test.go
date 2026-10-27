//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

type billingTenantFixture struct {
	authorityID       uuid.UUID
	offeringID        uuid.UUID
	offeringVersionID uuid.UUID
	priceVersionID    uuid.UUID
	counterPeriodID   uuid.UUID
	gaugePeriodID     uuid.UUID
	now               time.Time
}

func prepareBillingTenantFixture(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	globalAuthorityID uuid.UUID,
) billingTenantFixture {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	fixture := billingTenantFixture{
		authorityID: billingV7(t), counterPeriodID: billingV7(t),
		gaugePeriodID: billingV7(t), now: now,
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1($1,$2,1,1,$3)
`, fixture.authorityID, uuid.MustParse(billingTenantID), now); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	result := offeringResult{}
	fixture.offeringID = billingV7(t)
	fixture.offeringVersionID = billingV7(t)
	fixture.priceVersionID = billingV7(t)
	err = transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_offering_v1(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
)`, globalAuthorityID, billingV7(t), fixture.offeringID, fixture.offeringVersionID,
		uuid.New(), "usage-fixture", billingDigest("usage-fixture"),
		"usage-fixture", "Usage Fixture", false, true, 0, 0, int16(1), int16(2),
		now.Add(-time.Hour), nil, now,
	).Scan(&result.offeringID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
SELECT platform.billing_add_offering_price_v1($1,$2,$3,$4,$5,$6,$7,$8)
`, globalAuthorityID, fixture.offeringVersionID, fixture.priceVersionID,
			"monthly", int64(1000), "USD", int64(1), now)
	}
	for _, grant := range []struct {
		key   string
		value int64
	}{{"forms.submissions", 5}, {"documents.storage_bytes", 100}} {
		if err == nil {
			_, err = transaction.ExecContext(ctx, `
SELECT platform.billing_add_offering_grant_v1($1,$2,$3,'integer',NULL,$4,false,$5)
`, globalAuthorityID, fixture.offeringVersionID, grant.key, grant.value, now)
		}
	}
	if err == nil {
		err = seedBillingUsageRows(ctx, transaction, fixture)
	}
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("prepare Billing tenant fixture: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func seedBillingUsageRows(
	ctx context.Context,
	transaction *sql.Tx,
	fixture billingTenantFixture,
) error {
	tenantID := uuid.MustParse(billingTenantID)
	monthStart := time.Date(
		fixture.now.Year(), fixture.now.Month(), 1, 0, 0, 0, 0, time.UTC,
	)
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.billing_tenant_base_offerings(
 id,authority_id,tenant_id,offering_version_id,effective_start,effective_end,
 source_operation_id,created_at
) VALUES($1,$2,$3,$4,$5,NULL,$6,$5)
`, billingV7Must(), fixture.authorityID, tenantID, fixture.offeringVersionID,
		fixture.now.Add(-time.Hour), billingV7Must())
	if err != nil {
		return err
	}
	for _, period := range []struct {
		id, definition    uuid.UUID
		key, meter, reset string
		start, end        time.Time
	}{
		{fixture.counterPeriodID, uuid.MustParse("01900000-0000-7000-8000-000000000005"),
			"forms.submissions", "counter", "calendar_month_utc",
			monthStart, monthStart.AddDate(0, 1, 0)},
		{fixture.gaugePeriodID, uuid.MustParse("01900000-0000-7000-8000-000000000003"),
			"documents.storage_bytes", "gauge", "never",
			fixture.now.Add(-24 * time.Hour), fixture.now.AddDate(100, 0, 0)},
	} {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_meter_periods(
 id,authority_id,tenant_id,entitlement_definition_id,entitlement_key,meter_kind,
 reset_policy,period_start,period_end,state,closed_at,revision
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'open',NULL,1)
`, period.id, fixture.authorityID, tenantID, period.definition, period.key,
			period.meter, period.reset, period.start, period.end)
		if err != nil {
			return err
		}
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_usage_totals(
 authority_id,tenant_id,meter_period_id,entitlement_key,meter_kind,
 committed_total,reserved_total,gauge_total,revision,updated_at
) VALUES($1,$2,$3,$4,$5,0,0,0,1,$6)
`, fixture.authorityID, tenantID, period.id, period.key, period.meter,
			fixture.now.Add(-time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func loadBillingAdapterSQL(t *testing.T, adapter string, name string) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(
		workingDirectory, "..", adapter, "sql", "platform", name,
	))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
