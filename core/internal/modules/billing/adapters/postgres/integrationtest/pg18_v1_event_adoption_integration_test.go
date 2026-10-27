//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertBillingV1EventArchive(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	var eventID, tenantID, subscriptionID uuid.UUID
	var eventType, currency, description, metadata string
	var amount int64
	err := connection.QueryRowContext(ctx, `
SELECT source_event_id,tenant_id,source_subscription_id,event_type,
       amount_cents,currency,description,metadata::text
FROM platform.billing_v1_event_archive
`).Scan(&eventID, &tenantID, &subscriptionID, &eventType,
		&amount, &currency, &description, &metadata)
	if err != nil {
		t.Fatalf("read adopted Billing V1 event: %v", err)
	}
	if eventID.String() != "7b000000-0000-4000-8000-00000000007b" ||
		tenantID.String() != billingTenantID ||
		subscriptionID.String() != "75000000-0000-4000-8000-000000000075" ||
		eventType != "payment_succeeded" || amount != 900 || currency != "USD" ||
		description != "paid" || metadata != `{"receipt": "R-1"}` {
		t.Fatalf("adopted event=%s/%s/%s %s %d %s %q %s",
			eventID, tenantID, subscriptionID, eventType,
			amount, currency, description, metadata)
	}
	if _, err = connection.ExecContext(ctx, `
UPDATE platform.billing_v1_event_archive SET description='tampered'
`); err == nil {
		t.Fatal("immutable Billing V1 event archive accepted update")
	}
}

func assertBillingV1CompatibilityReports(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, statement := range []string{
		"UPDATE platform.cutover_groups SET state='v2',writer='v2'",
		"UPDATE platform.domain_cutovers SET state='v2',writer='v2'",
	} {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("activate retained Billing report fixture: %v", err)
		}
	}
	if _, err := database.ExecContext(ctx, `
UPDATE platform.tenants SET status='SUSPENDED',deleted_at=$1 WHERE id=$2
`, time.Now().UTC(), uuid.MustParse(billingTenantID)); err != nil {
		t.Fatalf("suspend retained Billing report tenant: %v", err)
	}
	if _, err := database.ExecContext(ctx, `
GRANT harpy_runtime TO harpy_v2_runtime WITH INHERIT FALSE, SET TRUE
`); err != nil {
		t.Fatalf("grant Billing report runtime role: %v", err)
	}
	connection := billingRuntimeConnection(t, ctx, database)
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	var eventID, tenantID, subscriptionID uuid.UUID
	var eventType, currency, description, metadata string
	var amount int64
	err := connection.QueryRowContext(ctx, `
SELECT id,tenant_id,subscription_id,event_type,amount_cents,currency,
       description,metadata::text
FROM platform.billing_v1_event_report_query_v1
WHERE id='7b000000-0000-4000-8000-00000000007b'::uuid
`).Scan(&eventID, &tenantID, &subscriptionID, &eventType,
		&amount, &currency, &description, &metadata)
	if err != nil || tenantID.String() != billingTenantID ||
		subscriptionID.String() != "75000000-0000-4000-8000-000000000075" ||
		eventType != "payment_succeeded" || amount != 900 || currency != "USD" ||
		description != "paid" || metadata != `{"receipt": "R-1"}` {
		t.Fatalf("retained event report=%s/%s/%s %s %d %s %q %s err=%v",
			eventID, tenantID, subscriptionID, eventType, amount,
			currency, description, metadata, err)
	}
	assertBillingV1SubscriptionReport(t, ctx, connection)
	if _, err = connection.ExecContext(ctx,
		"SELECT count(*) FROM platform.billing_v1_event_archive"); billingPGCode(err) != "42501" {
		t.Fatalf("runtime archive read code=%q err=%v", billingPGCode(err), err)
	}
	for _, statement := range []string{
		"UPDATE platform.cutover_groups SET state='shadow',writer='v1'",
		"UPDATE platform.domain_cutovers SET state='shadow',writer='v1'",
	} {
		if _, err = database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("restore Billing V1 adoption cutover: %v", err)
		}
	}
	if _, err = database.ExecContext(ctx, `
UPDATE platform.tenants SET status='ACTIVE',deleted_at=NULL WHERE id=$1
`, uuid.MustParse(billingTenantID)); err != nil {
		t.Fatalf("restore Billing V1 adoption tenant: %v", err)
	}
}

func assertBillingV1SubscriptionReport(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	var legacyID, legacyPlanID uuid.UUID
	var effective int64
	var currency string
	var registry int64
	var createdAt, updatedAt time.Time
	err := connection.QueryRowContext(ctx, `
SELECT legacy_id,legacy_plan_id,effective_price_minor_units,
       effective_price_currency,effective_price_registry_version,
       created_at,updated_at
FROM platform.billing_subscription_report_query_v2
`).Scan(&legacyID, &legacyPlanID, &effective, &currency,
		&registry, &createdAt, &updatedAt)
	if err != nil || legacyID.String() != "75000000-0000-4000-8000-000000000075" ||
		legacyPlanID.String() != "72000000-0000-4000-8000-000000000072" ||
		effective != 900 || currency != "USD" || registry != 1 ||
		createdAt.IsZero() || updatedAt.IsZero() || !updatedAt.After(createdAt) {
		t.Fatalf("subscription report=%s/%s %d %s/%d %s/%s err=%v",
			legacyID, legacyPlanID, effective, currency, registry,
			createdAt, updatedAt, err)
	}
}
