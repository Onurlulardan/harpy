//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestProviderUnknownReconciliationRouteOnPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := prepareProviderPG18(t, ctx)
	fixture := prepareProviderPGFixture(t, ctx, database)
	initialAttempt := createProviderPGCommand(t, ctx, &fixture)
	makeProviderPGUnknown(t, ctx, fixture.connection, fixture)

	nextFinalizer, reconciliationID := billingPGV7(t), billingPGV7(t)
	transaction, err := fixture.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	result := finalizeProviderPG(t, ctx, transaction, fixture,
		fixture.finalizerID, initialAttempt, nextFinalizer, reconciliationID)
	if result.decision != "reconcile" || result.state != "reconcile_required" ||
		result.activeID != reconciliationID || result.finalizerID != nextFinalizer ||
		result.activeID == result.finalizerID {
		t.Fatalf("unknown command finalization = %#v", result)
	}
	insertProviderPGReconciliation(t, ctx, transaction, fixture, reconciliationID)
	nextAttempt := insertProviderPGFinalizer(
		t, ctx, transaction, fixture, nextFinalizer,
	)
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}

	makeProviderPGJobUnknown(t, ctx, fixture.connection, fixture, reconciliationID)
	terminal := finalizeProviderPG(t, ctx, fixture.connection, fixture,
		nextFinalizer, nextAttempt, billingPGV7(t), billingPGV7(t))
	if terminal.decision != "reconcile_unknown" ||
		terminal.state != "reconcile_required" ||
		terminal.activeID != reconciliationID ||
		terminal.finalizerID != nextFinalizer || terminal.revision != 3 {
		t.Fatalf("reconciliation unknown finalization = %#v", terminal)
	}
}
