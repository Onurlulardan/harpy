//go:build integration

package integrationtest

import (
	"testing"

	"github.com/google/uuid"
)

func TestBillingRetentionTenantIsolationAndBoundariesOnPG18(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	fixture := prepareRetentionFixture(t, ctx)
	if _, err := fixture.connection.ExecContext(
		ctx, "SET ROLE harpy_runtime",
	); err != nil {
		t.Fatal(err)
	}
	_, err := fixture.connection.ExecContext(ctx, `
DELETE FROM platform.billing_webhook_receipts WHERE id=$1
`, fixture.dueA)
	if billingPGCode(err) != "42501" {
		t.Fatalf("runtime direct DML code=%q err=%v", billingPGCode(err), err)
	}
	result := executeRetention(
		t, ctx, fixture.connection, fixture.authorityA, fixture.tenantA,
		billingV7(t), billingV7(t), billingV7(t),
		"billing-retention://tenant-a", retentionDigest("a"),
	)
	if result.redacted != 1 || result.raw != 0 || result.held != 0 ||
		result.checkpoint != 1 {
		t.Fatalf("retention result=%+v", result)
	}
	if _, err = fixture.connection.ExecContext(
		ctx, "SET ROLE harpy_owner",
	); err != nil {
		t.Fatal(err)
	}
	var dueANull, futureANull, dueBNull bool
	for _, selected := range []struct {
		id     uuid.UUID
		target *bool
	}{
		{fixture.dueA, &dueANull},
		{fixture.futureA, &futureANull},
		{fixture.dueB, &dueBNull},
	} {
		if err = fixture.connection.QueryRowContext(ctx, `
SELECT encrypted_body IS NULL
FROM platform.billing_webhook_receipts WHERE id=$1
`, selected.id).Scan(selected.target); err != nil {
			t.Fatal(err)
		}
	}
	if !dueANull || futureANull || dueBNull {
		t.Fatalf(
			"redaction boundary due_a=%t future_a=%t due_b=%t",
			dueANull, futureANull, dueBNull,
		)
	}
	if _, err = fixture.connection.ExecContext(
		ctx, "SET ROLE harpy_runtime",
	); err != nil {
		t.Fatal(err)
	}
	err = fixture.connection.QueryRowContext(ctx, `
SELECT checkpoint_revision
FROM platform.billing_execute_retention_cycle_v1($1,$2,$3,$4,$5,$6,$7)
`, fixture.authorityA, fixture.tenantB, billingV7(t), billingV7(t),
		billingV7(t), "billing-retention://wrong-tenant", retentionDigest("b"),
	).Scan(new(int64))
	if billingPGCode(err) != "22023" {
		t.Fatalf("cross-tenant cycle code=%q err=%v", billingPGCode(err), err)
	}
}
