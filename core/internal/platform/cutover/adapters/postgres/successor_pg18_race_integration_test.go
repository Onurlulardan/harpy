//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type successorFinalizeResult struct {
	operation successorOperation
	err       error
}

func assertBillingFormsConcurrentCompletion(
	t *testing.T, ctx context.Context,
	database cutoverIntegrationDatabase, index int,
) {
	t.Helper()
	fixture := seedSuccessorFixture(t, ctx, database, index, "billing_forms",
		[]string{"billing", "forms"}, 6)
	seedSuccessorZeroInventory(t, ctx, database, fixture)
	for _, operation := range fixture.operations {
		allowSuccessorReceipt(t, ctx, database.admin, operation)
	}

	blocker, err := pgx.Connect(ctx, database.dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Close(ctx)
	transaction, err := blocker.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	if _, err = transaction.Exec(ctx, `
SELECT pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
    $1||':'||$2::text||':'||$3||':'||$4::bigint::text,0))`,
		fixture.scopeKind, fixture.scopeID, fixture.groupID,
		fixture.generation); err != nil {
		t.Fatal(err)
	}

	first := connectFinalizerSession(t, ctx, database.dsn)
	defer first.Close(ctx)
	second := connectFinalizerSession(t, ctx, database.dsn)
	defer second.Close(ctx)
	firstPID := successorBackendPID(t, ctx, first)
	secondPID := successorBackendPID(t, ctx, second)
	results := make(chan successorFinalizeResult, 2)
	go runSuccessorFinalize(ctx, first, fixture.operations[0], results)
	go runSuccessorFinalize(ctx, second, fixture.operations[1], results)
	waitForSuccessorLock(t, ctx, database.admin, firstPID)
	waitForSuccessorLock(t, ctx, database.admin, secondPID)
	assertSuccessorRoute(t, ctx, database.admin, fixture,
		"quiescing", "none", 6, 0, 0)
	if err = transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		select {
		case result := <-results:
			if result.err != nil {
				t.Fatalf("concurrent %s completion: %v",
					result.operation.domain, result.err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("billing forms concurrent completion did not resume")
		}
	}
	assertSuccessorRoute(t, ctx, database.admin, fixture,
		"v2", "v2", 7, 1, 3)
	triggering := successorTriggeringOperation(t, ctx, database.admin, fixture)
	assertSuccessorAtomicXID(t, ctx, database.admin, fixture, triggering)
}

func successorBackendPID(
	t *testing.T, ctx context.Context, connection *pgx.Conn,
) int {
	t.Helper()
	var backendPID int
	if err := connection.QueryRow(ctx,
		`SELECT pg_catalog.pg_backend_pid()`).Scan(&backendPID); err != nil {
		t.Fatal(err)
	}
	return backendPID
}

func runSuccessorFinalize(
	ctx context.Context, connection *pgx.Conn, operation successorOperation,
	results chan<- successorFinalizeResult,
) {
	results <- successorFinalizeResult{
		operation: operation, err: finalizeSuccessor(ctx, connection, operation),
	}
}

func waitForSuccessorLock(
	t *testing.T, ctx context.Context, admin *pgx.Conn, backendPID int,
) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var waitType string
		err := admin.QueryRow(ctx, `
SELECT COALESCE(wait_event_type,'') FROM pg_catalog.pg_stat_activity
WHERE pid=$1`, backendPID).Scan(&waitType)
		if err != nil {
			t.Fatal(err)
		}
		if waitType == "Lock" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("successor completion did not reach the serialization lock")
}

func successorTriggeringOperation(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	fixture successorFixture,
) successorOperation {
	t.Helper()
	var operationID string
	err := admin.QueryRow(ctx, `
SELECT receipt.operation_id::text
FROM cutover_authority.finalization_receipts AS receipt
JOIN platform.successor_authority_test_calls AS call
  ON call.call_xid::text=receipt.xmin::text
WHERE receipt.operation_id=ANY($1::uuid[])`, successorOperationIDs(fixture)).
		Scan(&operationID)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range fixture.operations {
		if operation.operationID == operationID {
			return operation
		}
	}
	t.Fatalf("successor triggering operation %s is unknown", operationID)
	return successorOperation{}
}
