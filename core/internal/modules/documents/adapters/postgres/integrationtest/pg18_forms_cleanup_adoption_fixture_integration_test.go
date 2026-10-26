//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

func prepareDocumentsFormsCleanupAdoption(t *testing.T, ctx context.Context,
	tx *sql.Tx, schema, operationID, definitionID, cursorID, linkID string,
	now time.Time,
) {
	t.Helper()
	assertDocumentsFormsCleanupWaitsForAdoption(t, ctx, tx,
		operationID, linkID, now)
	seedDocumentsFormsCleanupAdoption(t, ctx, tx, schema,
		definitionID, cursorID, now)
}

func assertDocumentsFormsCleanupWaitsForAdoption(t *testing.T,
	ctx context.Context, tx *sql.Tx, operationID, linkID string, now time.Time,
) {
	t.Helper()
	execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"prepare.sql",
		operationID, now)
	var state, linkState string
	var targets int64
	err := tx.QueryRowContext(ctx, `SELECT cleanup.lifecycle_state,
 link.lifecycle_state,(SELECT count(*) FROM documents_forms_cleanup_targets
 WHERE operation_id=cleanup.operation_id)
 FROM documents_forms_cleanup_operations AS cleanup
 JOIN documents_links AS link ON link.id=$2
 WHERE cleanup.operation_id=$1`, operationID, linkID).
		Scan(&state, &linkState, &targets)
	if err != nil || state != "REQUESTED" || linkState != "ACTIVE" || targets != 0 {
		t.Fatalf("cleanup crossed incomplete adoption = %s/%s/%d/%v",
			state, linkState, targets, err)
	}
}

func seedDocumentsFormsCleanupAdoption(t *testing.T, ctx context.Context,
	tx *sql.Tx, schema, definitionID, cursorID string, now time.Time,
) {
	t.Helper()
	source := sha256.Sum256([]byte(schema + ":forms-adoption-source"))
	catalog := sha256.Sum256([]byte(schema + ":forms-adoption-catalog"))
	sourceReceipt := sha256.Sum256([]byte(schema + ":forms-adoption-source-receipt"))
	receipt := sha256.Sum256([]byte(schema + ":forms-adoption-receipt"))
	_, err := tx.ExecContext(ctx, `INSERT INTO documents_form_record_adoption_receipts(
 operation_id,definition_id,reference_epoch,cursor_record_id,record_count,link_count,
 source_catalog_sha256,link_catalog_sha256,lifecycle_state,source_receipt_sha256,
 receipt_sha256,started_at,updated_at,completed_at)
 VALUES ('018f0000-0000-7000-8000-000000000813',$1,2,$2,1,1,
 $3,$4,'SUCCEEDED',$5,$6,$7,$7,$7)`, definitionID, cursorID,
		source[:], catalog[:], sourceReceipt[:], receipt[:], now)
	if err != nil {
		t.Fatalf("seed completed Documents Forms adoption: %v", err)
	}
}
