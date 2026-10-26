//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

const documentsFormsCleanupSQLRoot = "adapters/postgres/formscleanup/sql/tenant/"

func assertDocumentsFormsCleanupSQL(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	const (
		item        = "018f0000-0000-7000-8000-000000000801"
		version     = "018f0000-0000-7000-8000-000000000802"
		link        = "018f0000-0000-7000-8000-000000000803"
		definition  = "018f0000-0000-7000-8000-000000000804"
		record      = "018f0000-0000-7000-8000-000000000805"
		cleanup     = "018f0000-0000-7000-8000-000000000806"
		purgeJob    = "018f0000-0000-7000-8000-000000000807"
		replayJob   = "018f0000-0000-7000-8000-00000000080f"
		deletion    = "018f0000-0000-7000-8000-000000000808"
		missingVer  = "018f0000-0000-7000-8000-000000000810"
		missingDel  = "018f0000-0000-7000-8000-000000000811"
		missingBill = "018f0000-0000-7000-8000-000000000812"
		purgeOp     = "018f0000-0000-7000-8000-000000000809"
		purgeAudit  = "018f0000-0000-7000-8000-00000000080a"
		finishOp    = "018f0000-0000-7000-8000-00000000080b"
		finishAud   = "018f0000-0000-7000-8000-00000000080c"
		lease       = "018f0000-0000-7000-8000-00000000080d"
		billing     = "018f0000-0000-7000-8000-00000000080e"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	plain := sha256.Sum256([]byte(schema + ":plain"))
	cipher := sha256.Sum256([]byte(schema + ":cipher"))
	withDocumentsSchema(t, ctx, database, schema, func(tx *sql.Tx) {
		_, err := tx.ExecContext(ctx, `INSERT INTO documents_items
 (id,folder_id,owner_id,current_version_id,lifecycle_state,row_version,created_at,updated_at)
 VALUES ($1,'018f0000-0000-7000-8000-000000000101',
 '22222222-2222-4222-8222-222222222222',$2,'ACTIVE',1,$3,$3)`,
			item, version, now)
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO documents_versions
 (id,item_id,version_number,lifecycle_state,original_name,declared_mime,detected_mime,
  file_kind,plaintext_bytes,ciphertext_bytes,plaintext_sha256,ciphertext_sha256,
  cipher_format,object_key,object_version,object_generation,key_id,wrapped_dek,
  scan_signature_revision,scanned_at,created_at)
 VALUES ($1,$2,1,'READY','form.pdf','application/pdf','application/pdf','PDF',
  7,117,$3,$4,'FCDOC1',$5,'opaque-forms-cleanup',1,'documents-kek-1',
  $6,'clamav-20260731',$7,$7)`,
				version, item, plain[:], cipher[:],
				"tenants/test/documents/"+item+"/versions/"+version+"/content",
				make([]byte, 115), now)
		}
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO documents_links
 (id,item_id,source_owner,source_kind,source_id,source_version_id,usage,
  reference_epoch,lifecycle_state,created_at)
 VALUES ($1,$2,'FORMS','form_record',$3,$4,'LIFECYCLE',2,'ACTIVE',$5)`,
				link, item, definition, record, now)
		}
		if err != nil {
			t.Fatal(err)
		}
		execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"create.sql",
			cleanup, definition, int64(2), now)
		prepareDocumentsFormsCleanupAdoption(t, ctx, tx, schema, cleanup,
			definition, record, link, now)
		execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"prepare.sql",
			cleanup, now)
		assertDocumentsFormsCleanupEnqueueReplay(t, ctx, tx,
			cleanup, item, purgeJob, replayJob, now)
		execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"finalize.sql",
			cleanup, now)
		var state string
		var receipt []byte
		if err = tx.QueryRowContext(ctx, `SELECT lifecycle_state,receipt_sha256
 FROM documents_forms_cleanup_operations WHERE operation_id=$1`,
			cleanup).Scan(&state, &receipt); err != nil || state != "RUNNING" ||
			receipt != nil {
			t.Fatalf("%s premature receipt = %s/%x, %v", schema, state, receipt, err)
		}
	})
	purgeAt := now.Add(2 * time.Microsecond)
	tasks := []byte(`[{"deletion_id":"` + deletion + `","version_id":"` + version + `"}]`)
	withDocumentsSchema(t, ctx, database, schema, func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsPurgeSQLRoot+"lock.sql", item, purgeAt)
		execDocumentsSQL(t, ctx, tx, documentsPurgeSQLRoot+"begin.sql",
			item, purgeAt, tasks, purgeOp, purgeAudit)
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/claim.sql",
			deletion, lease, purgeAt, purgeAt.Add(time.Minute))
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/begin_reconcile.sql",
			deletion, lease, int64(1))
		deleteHash := sha256.Sum256([]byte(schema + ":delete"))
		absenceHash := sha256.Sum256([]byte(schema + ":absent"))
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/confirm.sql",
			deletion, int64(1), "provider-forms-cleanup",
			deleteHash[:], absenceHash[:], billing, purgeAt)
		execDocumentsSQL(t, ctx, tx, documentsPurgeSQLRoot+"finish.sql",
			item, purgeAt, finishOp, finishAud)
		insertDocumentsMissingDeletionVersion(t, ctx, tx, schema,
			item, missingVer, purgeAt)
		execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"finalize.sql",
			cleanup, purgeAt)
		assertDocumentsFormsCleanupMissingDeletionBlocked(t, ctx, tx, cleanup)
		insertDocumentsConfirmedDeletion(t, ctx, tx, schema,
			missingDel, missingVer, missingBill, purgeAt)
		execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"finalize.sql",
			cleanup, purgeAt)
	})
	assertDocumentsFormsCleanupReceipt(t, ctx, database, schema, cleanup, definition)
}

func assertDocumentsFormsCleanupReceipt(t *testing.T, ctx context.Context,
	database *sql.DB, schema, operationID, definitionID string,
) {
	t.Helper()
	var state, storedDefinition, linkState, itemState, versionState, deletionState string
	var links, items, versions, bytes int64
	var zero, receipt []byte
	err := database.QueryRowContext(ctx, `SELECT cleanup.lifecycle_state,
 cleanup.definition_id,cleanup.link_count,cleanup.item_count,cleanup.version_count,
 cleanup.ciphertext_bytes,cleanup.zero_blocker_sha256,cleanup.receipt_sha256,
 link.lifecycle_state,item.lifecycle_state,version.lifecycle_state,deletion.lifecycle_state
 FROM `+schema+`.documents_forms_cleanup_operations AS cleanup
 JOIN `+schema+`.documents_forms_cleanup_targets AS target
   ON target.operation_id=cleanup.operation_id
 JOIN `+schema+`.documents_links AS link ON link.item_id=target.item_id
 JOIN `+schema+`.documents_items AS item ON item.id=target.item_id
 JOIN `+schema+`.documents_versions AS version ON version.item_id=item.id
 JOIN `+schema+`.documents_blob_deletions AS deletion ON deletion.target_id=version.id
 WHERE cleanup.operation_id=$1`, operationID).Scan(
		&state, &storedDefinition, &links, &items, &versions, &bytes, &zero, &receipt,
		&linkState, &itemState, &versionState, &deletionState)
	if err != nil || state != "SUCCEEDED" || storedDefinition != definitionID ||
		links != 1 || items != 1 || versions != 2 || bytes != 236 ||
		len(zero) != 32 || len(receipt) != 32 || linkState != "RELEASED" ||
		itemState != "PURGED" || versionState != "PURGED" || deletionState != "CONFIRMED" {
		t.Fatalf("%s cleanup evidence invalid: %s %d/%d/%d/%d %x/%x %s/%s/%s/%s %v",
			schema, state, links, items, versions, bytes, zero, receipt,
			linkState, itemState, versionState, deletionState, err)
	}
	assertDocumentsFormsCleanupReplay(t, ctx, database, schema, operationID)
	assertDocumentsFormsCleanupTamperRejected(t, ctx, database, schema, operationID)
}

func assertDocumentsFormsCleanupReplay(t *testing.T, ctx context.Context,
	database *sql.DB, schema, operationID string,
) {
	t.Helper()
	withDocumentsSchema(t, ctx, database, schema, func(tx *sql.Tx) {
		var beforeReceipt, afterReceipt []byte
		var beforeCompleted, afterCompleted time.Time
		if err := tx.QueryRowContext(ctx, `SELECT receipt_sha256,completed_at
 FROM documents_forms_cleanup_operations WHERE operation_id=$1`,
			operationID).Scan(&beforeReceipt, &beforeCompleted); err != nil {
			t.Fatal(err)
		}
		execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"finalize.sql",
			operationID, beforeCompleted.Add(time.Minute))
		if err := tx.QueryRowContext(ctx, `SELECT receipt_sha256,completed_at
 FROM documents_forms_cleanup_operations WHERE operation_id=$1`,
			operationID).Scan(&afterReceipt, &afterCompleted); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(beforeReceipt, afterReceipt) ||
			!beforeCompleted.Equal(afterCompleted) {
			t.Fatal("Documents Forms cleanup replay changed immutable receipt")
		}
	})
}

func assertDocumentsFormsCleanupTamperRejected(t *testing.T, ctx context.Context,
	database *sql.DB, schema, operationID string,
) {
	t.Helper()
	withDocumentsSchema(t, ctx, database, schema, func(tx *sql.Tx) {
		if _, err := tx.ExecContext(ctx, "SAVEPOINT cleanup_tamper"); err != nil {
			t.Fatal(err)
		}
		_, err := tx.ExecContext(ctx, `UPDATE documents_forms_cleanup_operations
 SET receipt_sha256=$2 WHERE operation_id=$1`, operationID, make([]byte, 32))
		if err == nil {
			t.Fatal("Documents Forms cleanup receipt tamper was accepted")
		}
		if _, err = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT cleanup_tamper"); err != nil {
			t.Fatal(err)
		}
	})
}
