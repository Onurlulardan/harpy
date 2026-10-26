//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type normalOwnerFinalizationSnapshot struct {
	operationID, runID string
	state              string
	blockedReason      *string
	updatedAt          time.Time
	catalogDocument    []byte
	catalogSHA256      string
}

func driftNormalOwnerFinalizationRows(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) normalOwnerFinalizationSnapshot {
	t.Helper()
	snapshot := normalOwnerFinalizationSnapshot{
		operationID: evidence.operationID, runID: evidence.runID,
	}
	withNormalOwnerGuardsDisabled(t, ctx, database, func(tx pgx.Tx) {
		err := tx.QueryRow(ctx, `SELECT state,blocked_reason,updated_at
FROM platform.cutover_operations WHERE operation_id=$1 AND run_id=$2`,
			evidence.operationID, evidence.runID).Scan(
			&snapshot.state, &snapshot.blockedReason, &snapshot.updatedAt)
		if err != nil {
			t.Fatalf("capture normal-owner operation: %v", err)
		}
		err = tx.QueryRow(ctx, `SELECT canonical_document,catalog_sha256
FROM platform.tenant_migration_requirement_catalogs WHERE migration_run_id=$1`,
			integrationReleaseRun).Scan(&snapshot.catalogDocument, &snapshot.catalogSHA256)
		if err != nil {
			t.Fatalf("capture normal-owner tenant catalog: %v", err)
		}
		execOne(t, ctx, tx, "drift normal-owner operation", `UPDATE platform.cutover_operations
SET state='blocked',blocked_reason='simulated normal-owner drift',
    updated_at=pg_catalog.clock_timestamp()
WHERE operation_id=$1 AND run_id=$2`, evidence.operationID, evidence.runID)
		execOne(t, ctx, tx, "drift normal-owner tenant catalog",
			`UPDATE platform.tenant_migration_requirement_catalogs
SET canonical_document=canonical_document||pg_catalog.convert_to(' ','UTF8'),
    catalog_sha256=pg_catalog.encode(pg_catalog.sha256(
        canonical_document||pg_catalog.convert_to(' ','UTF8')),'hex')
WHERE migration_run_id=$1`, integrationReleaseRun)
	})
	return snapshot
}

func restoreNormalOwnerFinalizationRows(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	snapshot normalOwnerFinalizationSnapshot,
) {
	t.Helper()
	withNormalOwnerGuardsDisabled(t, ctx, database, func(tx pgx.Tx) {
		execOne(t, ctx, tx, "restore normal-owner operation", `UPDATE platform.cutover_operations
SET state=$3,blocked_reason=$4,updated_at=$5
WHERE operation_id=$1 AND run_id=$2`, snapshot.operationID, snapshot.runID,
			snapshot.state, snapshot.blockedReason, snapshot.updatedAt)
		execOne(t, ctx, tx, "restore normal-owner tenant catalog",
			`UPDATE platform.tenant_migration_requirement_catalogs
SET canonical_document=$2,catalog_sha256=$3 WHERE migration_run_id=$1`,
			integrationReleaseRun, snapshot.catalogDocument, snapshot.catalogSHA256)
		var operationExact, catalogExact bool
		err := tx.QueryRow(ctx, `SELECT
EXISTS (SELECT 1 FROM platform.cutover_operations WHERE operation_id=$1 AND run_id=$2
 AND state=$3 AND blocked_reason IS NOT DISTINCT FROM $4 AND updated_at=$5),
EXISTS (SELECT 1 FROM platform.tenant_migration_requirement_catalogs
 WHERE migration_run_id=$6 AND canonical_document=$7 AND catalog_sha256=$8)`,
			snapshot.operationID, snapshot.runID, snapshot.state, snapshot.blockedReason,
			snapshot.updatedAt, integrationReleaseRun, snapshot.catalogDocument,
			snapshot.catalogSHA256).Scan(&operationExact, &catalogExact)
		if err != nil || !operationExact || !catalogExact {
			t.Fatalf("normal-owner restore: operation=%t catalog=%t err=%v",
				operationExact, catalogExact, err)
		}
	})
}

func withNormalOwnerGuardsDisabled(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	work func(pgx.Tx),
) {
	t.Helper()
	migrator := openCutoverMigrator(t, ctx, database.dsn)
	defer migrator.Close(ctx)
	tx, err := migrator.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SET LOCAL ROLE harpy_owner`); err != nil {
		t.Fatalf("enter normal-owner role: %v", err)
	}
	var sessionUser, currentUser string
	err = tx.QueryRow(ctx, `SELECT SESSION_USER,CURRENT_USER`).Scan(&sessionUser, &currentUser)
	if err != nil || sessionUser != "harpy_migrator" || currentUser != "harpy_owner" {
		t.Fatalf("normal-owner identity: session=%s current=%s err=%v",
			sessionUser, currentUser, err)
	}
	tables := []string{"cutover_operations", "tenant_migration_requirement_catalogs"}
	triggerModes := captureUserTriggerModes(t, ctx, tx, "platform", tables)
	for _, table := range tables {
		if _, err = tx.Exec(ctx, `ALTER TABLE platform.`+pgx.Identifier{table}.Sanitize()+
			` DISABLE TRIGGER USER`); err != nil {
			t.Fatalf("disable %s normal-owner guards: %v", table, err)
		}
	}
	work(tx)
	restoreUserTriggerModes(t, ctx, tx, "platform", triggerModes)
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit normal-owner mutation: %v", err)
	}
}

func execOne(
	t *testing.T, ctx context.Context, tx pgx.Tx, label, statement string, args ...any,
) {
	t.Helper()
	tag, err := tx.Exec(ctx, statement, args...)
	if err != nil || tag.RowsAffected() != 1 {
		t.Fatalf("%s: rows=%d err=%v", label, tag.RowsAffected(), err)
	}
}
