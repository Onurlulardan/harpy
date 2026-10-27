//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

func insertDocumentsMissingDeletionVersion(t *testing.T, ctx context.Context,
	tx *sql.Tx, schema, itemID, versionID string, now time.Time,
) {
	t.Helper()
	plain := sha256.Sum256([]byte(schema + ":missing-plain"))
	cipher := sha256.Sum256([]byte(schema + ":missing-cipher"))
	receipt := sha256.Sum256([]byte(schema + ":missing-receipt"))
	_, err := tx.ExecContext(ctx, `INSERT INTO documents_versions
 (id,item_id,version_number,lifecycle_state,original_name,declared_mime,detected_mime,
  file_kind,plaintext_bytes,ciphertext_bytes,plaintext_sha256,ciphertext_sha256,
  cipher_format,object_key,object_version,object_generation,key_id,wrapped_dek,
  scan_signature_revision,scanned_at,delete_receipt_sha256,purged_at,created_at)
 VALUES ($1,$2,2,'PURGED','missing.pdf','application/pdf','application/pdf','PDF',
  9,119,$3,$4,'FCDOC1',$5,'opaque-missing-deletion',2,'documents-kek-1',
  $6,'clamav-20260731',$7,$8,$7,$7)`,
		versionID, itemID, plain[:], cipher[:],
		"tenants/test/documents/"+itemID+"/versions/"+versionID+"/content",
		make([]byte, 115), now, receipt[:])
	if err != nil {
		t.Fatal(err)
	}
}

func assertDocumentsFormsCleanupMissingDeletionBlocked(t *testing.T,
	ctx context.Context, tx *sql.Tx, operationID string,
) {
	t.Helper()
	var state string
	if err := tx.QueryRowContext(ctx, `SELECT lifecycle_state
 FROM documents_forms_cleanup_operations WHERE operation_id=$1`,
		operationID).Scan(&state); err != nil || state != "RUNNING" {
		t.Fatalf("missing deletion finalized cleanup: %s/%v", state, err)
	}
	if _, err := tx.ExecContext(ctx, "SAVEPOINT cleanup_missing_deletion"); err != nil {
		t.Fatal(err)
	}
	_, err := tx.ExecContext(ctx, `UPDATE documents_forms_cleanup_operations
 SET lifecycle_state='SUCCEEDED',version_count=2,ciphertext_bytes=236,
 zero_blocker_sha256=$2,receipt_sha256=$2,completed_at=$3
 WHERE operation_id=$1`, operationID, make([]byte, 32), time.Now().UTC())
	if err == nil {
		t.Fatal("cleanup trigger accepted a purged version without confirmed deletion")
	}
	if _, rollbackErr := tx.ExecContext(ctx,
		"ROLLBACK TO SAVEPOINT cleanup_missing_deletion"); rollbackErr != nil {
		t.Fatal(rollbackErr)
	}
}

func insertDocumentsConfirmedDeletion(t *testing.T, ctx context.Context,
	tx *sql.Tx, schema, deletionID, versionID, billingID string, now time.Time,
) {
	t.Helper()
	deleteHash := sha256.Sum256([]byte(schema + ":missing-delete"))
	absenceHash := sha256.Sum256([]byte(schema + ":missing-absence"))
	_, err := tx.ExecContext(ctx, `INSERT INTO documents_blob_deletions
 (id,target_kind,target_id,object_key,object_version,object_generation,
  plaintext_bytes,ciphertext_bytes,lifecycle_state,lease_generation,
  provider_request_id,delete_response_sha256,absence_probe_sha256,
  billing_adjustment_receipt_id,confirmed_at,created_at,updated_at)
 VALUES ($1,'VERSION',$2,$3,'opaque-missing-deletion',2,9,119,'CONFIRMED',1,
  'provider-missing-deletion',$4,$5,$6,$7,$7,$7)`,
		deletionID, versionID, "tenants/test/documents/missing/"+versionID,
		deleteHash[:], absenceHash[:], billingID, now)
	if err != nil {
		t.Fatal(err)
	}
}
