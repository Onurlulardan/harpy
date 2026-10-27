//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

type billingInvoiceFixture struct {
	billingSubscriptionFixture
	invoiceID     uuid.UUID
	invoiceLineID uuid.UUID
	policySHA256  []byte
}

func prepareBillingInvoiceFixture(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	subscription billingSubscriptionFixture,
) billingInvoiceFixture {
	t.Helper()
	fixture := billingInvoiceFixture{
		billingSubscriptionFixture: subscription,
		invoiceID:                  billingV7(t),
		invoiceLineID:              billingV7(t),
		policySHA256:               billingDigest("invoice-policy"),
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_invoices(
 id,authority_id,tenant_id,subscription_id,invoice_number,state,currency,
 currency_registry_version,currency_eligible,gross_minor_units,
 discount_minor_units,tax_minor_units,debit_adjustment_minor_units,
 credit_adjustment_minor_units,invoice_total_minor_units,settled_minor_units,
 amount_due_minor_units,credit_balance_minor_units,money_policy_version,
 policy_sha256,period_start,period_end,created_at,updated_at,finalized_at,revision
) VALUES(
 $1,$2,$3,$4,$5,'draft','USD',1,true,1000,100,180,20,50,1050,0,1050,0,
 'fc_money_v1',$6,$7,$8,$9,$9,NULL,1
)
`, fixture.invoiceID, fixture.authorityID, uuid.MustParse(billingTenantID),
		fixture.subscriptionID, "INV-"+fixture.invoiceID.String(), fixture.policySHA256,
		fixture.now.Add(-time.Hour), fixture.now.AddDate(0, 1, 0), fixture.now)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_invoice_lines(
 id,authority_id,tenant_id,invoice_id,line_sequence,line_kind,description,
 quantity,unit_minor_units,gross_minor_units,discount_minor_units,tax_minor_units,
 debit_adjustment_minor_units,credit_adjustment_minor_units,
 line_total_minor_units,currency,currency_registry_version,allocation_rank,
 money_policy_version,created_at
) VALUES($1,$2,$3,$4,1,'item','integration invoice',2,500,1000,100,180,20,50,
 1050,'USD',1,1,'fc_money_v1',$5)
`, fixture.invoiceLineID, fixture.authorityID, uuid.MustParse(billingTenantID),
			fixture.invoiceID, fixture.now)
	}
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("prepare Billing invoice fixture: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return fixture
}
