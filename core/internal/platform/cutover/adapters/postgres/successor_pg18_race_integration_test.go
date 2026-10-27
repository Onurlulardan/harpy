//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertBillingOnlySuccessorCompletion(
	t *testing.T, ctx context.Context,
	database cutoverIntegrationDatabase, index int,
) {
	t.Helper()
	fixture := seedSuccessorFixture(t, ctx, database, index, "billing",
		[]string{"billing"}, 6)
	seedSuccessorZeroInventory(t, ctx, database, fixture)
	allowSuccessorReceipt(t, ctx, database.admin, fixture.operations[0])
	finalizer := connectFinalizerSession(t, ctx, database.dsn)
	defer finalizer.Close(ctx)
	if err := finalizeSuccessor(ctx, finalizer, fixture.operations[0]); err != nil {
		t.Fatalf("billing successor finalize: %v", err)
	}
	assertSuccessorRoute(t, ctx, database.admin, fixture, "v2", "v2", 7, 1, 2)
	assertSuccessorAtomicXID(t, ctx, database.admin, fixture, fixture.operations[0])
	assertGlobalBillingSuccessorCall(t, ctx, database, fixture)
}
