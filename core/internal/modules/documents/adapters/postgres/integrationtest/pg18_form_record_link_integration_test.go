//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

const (
	formLinkActor   = "22222222-2222-4222-8222-222222222222"
	formLinkItem    = "018f0000-0000-7000-8000-000000000901"
	formLinkSource  = "018f0000-0000-7000-8000-000000000902"
	formLinkRecord  = "018f0000-0000-7000-8000-000000000903"
	formLinkVersion = "018f0000-0000-7000-8000-000000000940"
)

func assertDocumentsFormRecordLinks(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	plain := sha256.Sum256([]byte(schema + ":form-link-plain"))
	cipher := sha256.Sum256([]byte(schema + ":form-link-cipher"))
	withDocumentsSchema(t, ctx, database, schema, func(tx *sql.Tx) {
		_, err := tx.ExecContext(ctx, `INSERT INTO documents_items
 (id,folder_id,owner_id,current_version_id,lifecycle_state,reference_epoch,
  row_version,created_at,updated_at)
 VALUES ($1,'018f0000-0000-7000-8000-000000000101',$2,$3,'ACTIVE',9,1,$4,$4)`,
			formLinkItem, formLinkActor, formLinkVersion, now)
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO documents_versions
 (id,item_id,version_number,lifecycle_state,original_name,declared_mime,
  detected_mime,file_kind,plaintext_bytes,ciphertext_bytes,plaintext_sha256,
  ciphertext_sha256,cipher_format,object_key,object_version,object_generation,
  key_id,wrapped_dek,scan_signature_revision,scanned_at,created_at)
 VALUES ($1,$2,1,'READY','reference.pdf','application/pdf','application/pdf',
  'PDF',7,117,$3,$4,'FCDOC1',$5,'opaque-reference',1,'documents-kek-1',
  $6,'clamav-20260731',$7,$7)`,
				formLinkVersion, formLinkItem, plain[:], cipher[:],
				"tenants/test/documents/"+formLinkItem+"/versions/"+
					formLinkVersion+"/content",
				make([]byte, 115), now)
		}
		if err != nil {
			t.Fatal(err)
		}
	})
	registerSQL := readDocumentsLinkSQL(t, "register.sql")
	runConcurrentFormLinkRegisters(t, ctx, database, schema, registerSQL, now)

	var activeID string
	err := database.QueryRowContext(ctx, `SELECT id FROM `+schema+`.documents_links
 WHERE item_id=$1 AND source_owner='FORMS' AND source_kind='form_record'
   AND source_id=$2 AND source_version_id=$3 AND lifecycle_state='ACTIVE'`,
		formLinkItem, formLinkSource, formLinkRecord).Scan(&activeID)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(schema + ":release"))
	withDocumentsSchema(t, ctx, database, schema, func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsLinkSQLRoot+"release.sql",
			activeID, formLinkActor, "FORMS", "form_record", formLinkSource,
			formLinkRecord, now.Add(time.Microsecond),
			"018f0000-0000-7000-8000-000000000910", "form-link-release",
			hash[:], "018f0000-0000-7000-8000-000000000911")
		registerFormLinkSQL(t, ctx, tx, registerSQL,
			"018f0000-0000-7000-8000-000000000912",
			"018f0000-0000-7000-8000-000000000913",
			"018f0000-0000-7000-8000-000000000914",
			"form-link-readd", now.Add(2*time.Microsecond))
		registerFormLinkSQL(t, ctx, tx, registerSQL,
			"018f0000-0000-7000-8000-000000000912",
			"018f0000-0000-7000-8000-000000000913",
			"018f0000-0000-7000-8000-000000000914",
			"form-link-readd", now.Add(2*time.Microsecond))
	})
	var total, active, released int
	err = database.QueryRowContext(ctx, `SELECT count(*),
 count(*) FILTER (WHERE lifecycle_state='ACTIVE'),
 count(*) FILTER (WHERE lifecycle_state='RELEASED')
 FROM `+schema+`.documents_links
 WHERE item_id=$1 AND source_id=$2 AND source_version_id=$3`,
		formLinkItem, formLinkSource, formLinkRecord).Scan(&total, &active, &released)
	if err != nil || total != 2 || active != 1 || released != 1 {
		t.Fatalf("%s rebound rows=%d active=%d released=%d err=%v",
			schema, total, active, released, err)
	}
	var epoch int64
	err = database.QueryRowContext(ctx, `SELECT reference_epoch
 FROM `+schema+`.documents_links WHERE id=$1`,
		"018f0000-0000-7000-8000-000000000912").Scan(&epoch)
	if err != nil || epoch != 9 {
		t.Fatalf("%s rebound epoch=%d err=%v", schema, epoch, err)
	}
	assertFormLinkRollback(t, ctx, database, schema, registerSQL, now)
}
