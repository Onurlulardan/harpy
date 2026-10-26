//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func assertActorBridgeMaintenance(
	t *testing.T,
	ctx context.Context,
	database cutoverIntegrationDatabase,
	runtime *ActorBridgeDatabase,
	wantHits int64,
) {
	t.Helper()
	reconciliation, err := runtime.ReconcileHits(ctx, 100)
	if err != nil {
		var count int64
		var latest sql.NullTime
		rawErr := runtime.queryer.QueryRow(
			ctx, reconcileIAMV1ActorAssertionHitsSQL, 100,
		).Scan(&count, &latest)
		t.Fatalf("reconcile actor bridge hits: classified=%v raw=%v", err, rawErr)
	}
	if latest, exists := reconciliation.LatestHitAt(); reconciliation.AccountedCount() != wantHits || !exists || latest.IsZero() {
		t.Fatalf("actor bridge reconciliation = %d/%v/%v",
			reconciliation.AccountedCount(), latest, exists)
	}
	empty, err := runtime.ReconcileHits(ctx, 100)
	if err != nil || empty.AccountedCount() != 0 {
		t.Fatalf("empty actor bridge reconciliation = %+v, err=%v", empty, err)
	}

	var hitCount, revision, remaining int64
	if err := database.admin.QueryRow(ctx, `
SELECT hit_count, revision, remaining_scope_count
FROM platform.contract_retirements
WHERE contract_key='iam.v1_actor_assertion.v1'`).Scan(
		&hitCount, &revision, &remaining,
	); err != nil {
		t.Fatal(err)
	}
	if hitCount != wantHits || remaining != 2 {
		t.Fatalf("actor bridge retirement = hits %d, remaining %d", hitCount, remaining)
	}
	assertActorBridgeRetirementRefusesLiveTargets(t, ctx, database, revision)

	_, err = database.admin.Exec(ctx, `
UPDATE platform.iam_v1_actor_assertion_receipts
SET issued_at=statement_timestamp()-INTERVAL '2 days',
    expires_at=statement_timestamp()-INTERVAL '2 days'+INTERVAL '30 seconds',
    consumed_at=statement_timestamp()-INTERVAL '2 days'+INTERVAL '1 second'`)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := runtime.Purge(ctx, 100)
	if err != nil || deleted != wantHits {
		t.Fatalf("actor bridge purge = %d, err=%v", deleted, err)
	}
	var receipts int64
	if err := database.admin.QueryRow(ctx,
		"SELECT count(*) FROM platform.iam_v1_actor_assertion_receipts",
	).Scan(&receipts); err != nil || receipts != 0 {
		t.Fatalf("actor bridge receipts after purge = %d, err=%v", receipts, err)
	}
}

func assertActorBridgeRetirementRefusesLiveTargets(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase, revision int64,
) {
	t.Helper()
	migrator := openCutoverMigrator(t, ctx, database.dsn)
	defer migrator.Close(ctx)
	if _, err := migrator.Exec(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	var remaining, nextRevision int64
	if err := migrator.QueryRow(ctx,
		"SELECT * FROM platform.iam_v1_actor_bridge_reconcile_v1($1)", revision,
	).Scan(&remaining, &nextRevision); err != nil ||
		remaining != 2 || nextRevision != revision+1 {
		t.Fatalf("actor bridge scope reconciliation = %d/%d, err=%v",
			remaining, nextRevision, err)
	}
	_, err := migrator.Exec(ctx,
		"SELECT platform.iam_v1_actor_bridge_mark_eligible_v1($1)", nextRevision)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "P5607" {
		t.Fatalf("live actor bridge eligibility error = %v", err)
	}
}
