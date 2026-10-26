//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"testing"
)

const (
	deliveryClaimSQL = `SELECT * FROM
platform.platformauthority_delivery_claim_v1($1)`
	deliveryAwaitSQL = `SELECT
platform.platformauthority_delivery_await_import_v1($1,$2,$3)`
	deliverySucceedSQL = `SELECT
platform.platformauthority_delivery_succeed_v1($1,$2,$3)`
	deliveryRetrySQL = `SELECT
platform.platformauthority_delivery_retry_v1($1,$2,$3,$4,$5)`
)

type deliveryLease struct {
	operationID string
	generation  int64
	attempt     int
	maxAttempts int
}

func exerciseDeliveryQueue(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	ownerA := "99999999-9999-4999-8999-999999999991"
	ownerB := "99999999-9999-4999-8999-999999999992"
	first := claimDelivery(t, ctx, database, ownerA)
	if first.operationID != "55555555-5555-4555-8555-555555555555" ||
		first.attempt != 1 || first.maxAttempts != 8 {
		t.Fatalf("first delivery claim = %+v", first)
	}
	if _, err := database.ExecContext(
		ctx, deliveryAwaitSQL,
		first.operationID, ownerA, first.generation,
	); err != nil {
		t.Fatal(err)
	}
	completed := claimDelivery(t, ctx, database, ownerA)
	if completed.operationID != "66666666-6666-4666-8666-666666666666" {
		t.Fatalf("completed delivery claim = %+v", completed)
	}
	if _, err := database.ExecContext(
		ctx, deliverySucceedSQL,
		completed.operationID, ownerA, completed.generation,
	); err != nil {
		t.Fatal(err)
	}
	recoverable := claimDelivery(t, ctx, database, ownerA)
	if recoverable.operationID != "88888888-8888-4888-8888-888888888888" {
		t.Fatalf("recoverable delivery claim = %+v", recoverable)
	}
	if _, err := database.ExecContext(ctx, `
UPDATE platform.platformauthority_delivery_jobs
SET lease_until=clock_timestamp()-interval '1 second'
WHERE operation_id=$1`, recoverable.operationID); err != nil {
		t.Fatal(err)
	}
	recovered := claimDelivery(t, ctx, database, ownerB)
	if recovered.operationID != recoverable.operationID ||
		recovered.generation != recoverable.generation+1 ||
		recovered.attempt != recoverable.attempt+1 {
		t.Fatalf("recovered delivery claim = %+v", recovered)
	}
	if _, err := database.ExecContext(
		ctx, deliveryRetrySQL,
		recovered.operationID, ownerB, recovered.generation,
		"REGISTRY_UNAVAILABLE", false,
	); err != nil {
		t.Fatal(err)
	}
	var awaiting, succeeded int
	if err := database.QueryRowContext(ctx, `
SELECT count(*) FILTER (WHERE state='awaiting_import'),
       count(*) FILTER (WHERE state='succeeded')
FROM platform.platformauthority_delivery_jobs`).Scan(
		&awaiting, &succeeded,
	); err != nil || awaiting != 1 || succeeded != 1 {
		t.Fatalf("delivery states awaiting=%d succeeded=%d err=%v",
			awaiting, succeeded, err)
	}
}

func claimDelivery(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	owner string,
) deliveryLease {
	t.Helper()
	var lease deliveryLease
	if err := database.QueryRowContext(ctx, deliveryClaimSQL, owner).Scan(
		&lease.operationID, &lease.generation,
		&lease.attempt, &lease.maxAttempts,
	); err != nil {
		t.Fatal(err)
	}
	return lease
}
