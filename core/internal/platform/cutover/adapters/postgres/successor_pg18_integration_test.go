//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestCutoverSuccessorFinalizationAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			skipped := map[string]bool{
				"20260728110020_assert_runtime_snapshot_authorities_v1." +
					origin + ".sql": true,
			}
			database := startCutoverIntegrationDatabaseForOriginSkipping(
				t, ctx, origin, skipped)
			prepareIAMSystemResourcesAsyncSurfaces(t, ctx, database, origin)
			attachSuccessorReceiptHarness(t, ctx, database.admin)
			finalizer := connectFinalizerSession(t, ctx, database.dsn)
			defer finalizer.Close(ctx)

			index := 11
			for _, module := range []string{"analytics", "workflow"} {
				positive := seedSuccessorFixture(
					t, ctx, database, index, module, []string{module}, 6)
				seedSuccessorZeroInventory(t, ctx, database, positive)
				allowSuccessorReceipt(t, ctx, database.admin, positive.operations[0])
				if err := finalizeSuccessor(ctx, finalizer, positive.operations[0]); err != nil {
					t.Fatalf("%s positive successor finalize: %v", module, err)
				}
				assertSuccessorRoute(t, ctx, database.admin, positive,
					"v2", "v2", 7, 1, 2)
				assertSuccessorAtomicXID(
					t, ctx, database.admin, positive, positive.operations[0])
				if err := finalizeSuccessor(ctx, finalizer, positive.operations[0]); err != nil {
					t.Fatalf("%s successor replay: %v", module, err)
				}
				assertSuccessorRoute(t, ctx, database.admin, positive,
					"v2", "v2", 7, 1, 2)
				index++

				wrongReceipt := seedSuccessorFixture(
					t, ctx, database, index, module, []string{module}, 6)
				seedSuccessorZeroInventory(t, ctx, database, wrongReceipt)
				err := finalizeSuccessor(ctx, finalizer, wrongReceipt.operations[0])
				assertFinalizationError(
					t, err, "P5742", "cutover successor protected receipt differs")
				assertSuccessorRoute(t, ctx, database.admin, wrongReceipt,
					"quiescing", "none", 6, 0, 0)
				assertSuccessorOperationState(t, ctx, database.admin,
					wrongReceipt.operations[0], "quiescing", 0)
				index++

				wrongGeneration := seedSuccessorFixture(
					t, ctx, database, index, module, []string{module}, 5)
				seedSuccessorZeroInventory(t, ctx, database, wrongGeneration)
				allowSuccessorReceipt(
					t, ctx, database.admin, wrongGeneration.operations[0])
				err = finalizeSuccessor(ctx, finalizer, wrongGeneration.operations[0])
				assertFinalizationError(t, err, "P0002", "query returned no rows")
				assertSuccessorRoute(t, ctx, database.admin, wrongGeneration,
					"quiescing", "none", 6, 0, 0)
				assertSuccessorOperationState(t, ctx, database.admin,
					wrongGeneration.operations[0], "quiescing", 0)
				index++
			}

			billingForms := seedSuccessorFixture(
				t, ctx, database, index, "billing_forms",
				[]string{"billing", "forms"}, 6)
			seedSuccessorZeroInventory(t, ctx, database, billingForms)
			allowSuccessorReceipt(t, ctx, database.admin, billingForms.operations[0])
			if err := finalizeSuccessor(
				ctx, finalizer, billingForms.operations[0]); err != nil {
				t.Fatalf("billing partial successor finalize: %v", err)
			}
			assertSuccessorRoute(t, ctx, database.admin, billingForms,
				"quiescing", "none", 6, 0, 0)
			assertSuccessorOperationState(t, ctx, database.admin,
				billingForms.operations[0], "completed", 1)
			assertSuccessorOperationState(t, ctx, database.admin,
				billingForms.operations[1], "quiescing", 0)

			allowSuccessorReceipt(t, ctx, database.admin, billingForms.operations[1])
			if err := finalizeSuccessor(
				ctx, finalizer, billingForms.operations[1]); err != nil {
				t.Fatalf("forms completing successor finalize: %v", err)
			}
			assertSuccessorRoute(t, ctx, database.admin, billingForms,
				"v2", "v2", 7, 1, 3)
			assertSuccessorAtomicXID(
				t, ctx, database.admin, billingForms, billingForms.operations[1])
			assertBillingSuccessorCall(
				t, ctx, database, billingForms.operations[0], billingForms.operations[1])
			assertBillingFormsConcurrentCompletion(
				t, ctx, database, index+1)

			globalBilling := seedGlobalBillingSuccessorFixture(
				t, ctx, database, index+2, 6)
			seedSuccessorZeroInventory(t, ctx, database, globalBilling)
			allowSuccessorReceipt(
				t, ctx, database.admin, globalBilling.operations[0])
			if err := finalizeSuccessor(
				ctx, finalizer, globalBilling.operations[0]); err != nil {
				t.Fatalf("global Billing successor finalize: %v", err)
			}
			assertSuccessorRoute(t, ctx, database.admin, globalBilling,
				"v2", "v2", 7, 1, 2)
			assertSuccessorAtomicXID(
				t, ctx, database.admin, globalBilling,
				globalBilling.operations[0])
			assertGlobalBillingSuccessorCall(t, ctx, database, globalBilling)
			if err := finalizeSuccessor(
				ctx, finalizer, globalBilling.operations[0]); err != nil {
				t.Fatalf("global Billing successor replay: %v", err)
			}
			assertSuccessorRoute(t, ctx, database.admin, globalBilling,
				"v2", "v2", 7, 1, 2)

			globalWrongReceipt := seedGlobalBillingSuccessorFixture(
				t, ctx, database, index+3, 6)
			seedSuccessorZeroInventory(t, ctx, database, globalWrongReceipt)
			err := finalizeSuccessor(
				ctx, finalizer, globalWrongReceipt.operations[0])
			assertFinalizationError(
				t, err, "P5742", "cutover successor protected receipt differs")
			assertSuccessorRoute(t, ctx, database.admin, globalWrongReceipt,
				"quiescing", "none", 6, 0, 0)
			assertSuccessorOperationState(t, ctx, database.admin,
				globalWrongReceipt.operations[0], "quiescing", 0)
		})
	}
}

func assertBillingSuccessorCall(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	billing, forms successorOperation,
) {
	t.Helper()
	var companionOperation, companionRun, digest string
	err := database.admin.QueryRow(ctx, `
SELECT companion_operation_id::text,companion_run_id::text,acceptance_digest
FROM platform.successor_authority_test_calls
WHERE module='billing' AND operation_id=$1 AND run_id=$2`,
		billing.operationID, billing.runID).Scan(
		&companionOperation, &companionRun, &digest)
	if err != nil || companionOperation != forms.operationID ||
		companionRun != forms.runID || digest != successorAcceptanceDigest {
		t.Fatalf("billing successor call companion=%s/%s digest=%s err=%v",
			companionOperation, companionRun, digest, err)
	}
}

func assertGlobalBillingSuccessorCall(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	fixture successorFixture,
) {
	t.Helper()
	var exact bool
	err := database.admin.QueryRow(ctx, `
SELECT module='billing' AND companion_operation_id IS NULL
       AND companion_run_id IS NULL AND acceptance_digest=$2
FROM platform.successor_authority_test_calls
WHERE operation_id=$1`, fixture.operations[0].operationID,
		successorAcceptanceDigest).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("global Billing successor call exact=%t err=%v", exact, err)
	}
}
