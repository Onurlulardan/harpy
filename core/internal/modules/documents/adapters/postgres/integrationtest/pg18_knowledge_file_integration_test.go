//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

const knowledgeSQLRoot = "adapters/postgres/knowledgefile/sql/tenant/"

func TestDocumentsKnowledgeFileAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database := startDocumentsPG18(t, ctx)
	prepareDocumentsPG18(t, ctx, database)
	applyDocumentsMigrations(t, ctx, database)
	for _, schema := range []string{"documents_fresh_a", "documents_adopted_a"} {
		insertDocumentsFolder(t, ctx, database, schema,
			"018f0000-0000-7000-8000-000000000101")
		assertDocumentsKnowledgeFileSQL(t, ctx, database, schema)
	}
}

func assertDocumentsKnowledgeFileSQL(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	const (
		actor      = "22222222-2222-4222-8222-222222222222"
		item       = "018f0000-0000-7000-8000-000000000951"
		version    = "018f0000-0000-7000-8000-000000000952"
		upload     = "018f0000-0000-7000-8000-000000000953"
		reserve    = "018f0000-0000-7000-8000-000000000954"
		reference  = "018f0000-0000-7000-8000-000000000955"
		source     = "018f0000-0000-7000-8000-000000000956"
		lease      = "018f0000-0000-7000-8000-000000000957"
		receipt    = "018f0000-0000-7000-8000-000000000958"
		readyOp    = "018f0000-0000-7000-8000-000000000959"
		readyAudit = "018f0000-0000-7000-8000-000000000960"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	plain := sha256.Sum256([]byte(schema + ":knowledge-plain"))
	cipher := sha256.Sum256([]byte(schema + ":knowledge-cipher"))
	request := sha256.Sum256([]byte(schema + ":knowledge-request"))
	withDocumentsSchema(t, ctx, database, schema, func(tx *sql.Tx) {
		insertKnowledgeUpload(t, ctx, tx, upload, item, version, reserve,
			actor, request, plain, now)
		execDocumentsSQL(t, ctx, tx, knowledgeSQLRoot+"create.sql",
			upload, actor, item, version, int64(7), plain[:], "knowledge.txt",
			"text/plain", now.Add(time.Hour), now)
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"begin.sql",
			upload, actor, lease, receipt)
		objectKey := "tenant/" + schema + "/knowledge/" + version
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"record_effect.sql",
			upload, actor, int64(1), objectKey, "knowledge-v1", int64(1), int64(7),
			int64(117), plain[:], cipher[:], "text/plain", "txt", "documents-kek-1",
			make([]byte, 115), "clamav-v1", now)
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"prepare_finalize.sql",
			upload, actor, int64(1), "knowledge-v1")
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"finalize.sql",
			upload, actor, int64(1), objectKey, "knowledge-v1", int64(1), int64(7),
			int64(117), plain[:], cipher[:], "text/plain", "txt", "documents-kek-1",
			make([]byte, 115), "clamav-v1", now, readyOp, readyAudit)
		execDocumentsSQL(t, ctx, tx, knowledgeSQLRoot+"complete_stage.sql",
			upload, actor, now)
		execDocumentsSQL(t, ctx, tx, knowledgeSQLRoot+"prepare_adopt.sql",
			upload, actor, item, version, reference, source, now)
		_, err := tx.ExecContext(ctx, `INSERT INTO documents_links
 (id,item_id,source_owner,source_kind,source_id,source_version_id,usage,
  reference_epoch,lifecycle_state,created_at,released_at)
 VALUES($1,$2,'AI','knowledge_document',$3,$3,'LIFECYCLE',1,'ACTIVE',$4,NULL)`,
			reference, item, source, now)
		if err != nil {
			t.Fatal(err)
		}
		execDocumentsSQL(t, ctx, tx, knowledgeSQLRoot+"complete_adopt.sql",
			upload, actor, item, version, reference, source, now)
		execDocumentsSQL(t, ctx, tx, knowledgeSQLRoot+"verify_open.sql",
			actor, item, version)
		_, err = tx.ExecContext(ctx, `UPDATE documents_links
 SET lifecycle_state='RELEASED',released_at=$2 WHERE id=$1`, reference, now)
		if err != nil {
			t.Fatal(err)
		}
		execDocumentsSQL(t, ctx, tx, knowledgeSQLRoot+"complete_release.sql",
			item, reference, source, now)
	})
	var state string
	if err := database.QueryRowContext(ctx, `SELECT lifecycle_state FROM `+schema+
		`.documents_knowledge_file_leases WHERE lease_id=$1`, upload).Scan(&state); err != nil ||
		state != "RELEASED" {
		t.Fatalf("%s knowledge lease state=%s error=%v", schema, state, err)
	}
	assertKnowledgeExpirySQL(t, ctx, database, schema, now)
}

func insertKnowledgeUpload(t *testing.T, ctx context.Context, tx *sql.Tx,
	upload, item, version, reserve, actor string, request, plain [32]byte, now time.Time,
) {
	t.Helper()
	_, err := tx.ExecContext(ctx, `INSERT INTO documents_uploads
 (id,actor_id,document_id,candidate_version_id,folder_id,request_sha256,
  expected_content_sha256,expected_plaintext_bytes,ciphertext_upper_bound,
  original_name,declared_mime,billing_reservation_id,attempt_generation,
  lifecycle_state,cleanup_required,expires_at,created_at,updated_at)
 VALUES($1,$2,$3,$4,'018f0000-0000-7000-8000-000000000101',$5,$6,7,117,
  'knowledge.txt','text/plain',$7,0,'CREATED',false,$8,$9,$9)`,
		upload, actor, item, version, request[:], plain[:], reserve,
		now.Add(time.Hour), now)
	if err != nil {
		t.Fatal(err)
	}
}

func assertKnowledgeExpirySQL(t *testing.T, ctx context.Context,
	database *sql.DB, schema string, now time.Time,
) {
	t.Helper()
	const (
		actor   = "22222222-2222-4222-8222-222222222222"
		item    = "018f0000-0000-7000-8000-000000000961"
		version = "018f0000-0000-7000-8000-000000000962"
		upload  = "018f0000-0000-7000-8000-000000000963"
		reserve = "018f0000-0000-7000-8000-000000000964"
	)
	request, plain := sha256.Sum256([]byte(schema+":expired-request")),
		sha256.Sum256([]byte(schema+":expired-content"))
	cipher := sha256.Sum256([]byte(schema + ":expired-cipher"))
	withDocumentsSchema(t, ctx, database, schema, func(tx *sql.Tx) {
		_, err := tx.ExecContext(ctx, `INSERT INTO documents_items
	 (id,folder_id,owner_id,current_version_id,lifecycle_state,reference_epoch,
	  row_version,created_at,updated_at)
	 VALUES($1,'018f0000-0000-7000-8000-000000000101',$2,$3,'ACTIVE',1,7,$4,$4)`,
			item, actor, version, now.Add(-2*time.Hour))
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO documents_versions
	 (id,item_id,version_number,lifecycle_state,original_name,declared_mime,
	  detected_mime,file_kind,plaintext_bytes,ciphertext_bytes,plaintext_sha256,
	  ciphertext_sha256,cipher_format,object_key,object_version,object_generation,
	  key_id,wrapped_dek,scan_signature_revision,scanned_at,created_at)
	 VALUES($1,$2,1,'READY','knowledge.txt','text/plain','text/plain','TXT',7,117,
	  $3,$4,'FCDOC1',$5,'expired-v1',1,'documents-kek-1',$6,'clamav-v1',$7,$7)`,
				version, item, plain[:], cipher[:],
				"tenant/"+schema+"/knowledge-expired/"+version,
				make([]byte, 115), now.Add(-2*time.Hour))
		}
		if err != nil {
			t.Fatal(err)
		}
		insertKnowledgeUpload(t, ctx, tx, upload, item, version, reserve, actor,
			request, plain, now.Add(-2*time.Hour))
		execDocumentsSQL(t, ctx, tx, knowledgeSQLRoot+"create.sql", upload, actor,
			item, version, int64(7), plain[:], "knowledge.txt", "text/plain",
			now.Add(-time.Hour), now.Add(-2*time.Hour))
		execDocumentsSQL(t, ctx, tx, knowledgeSQLRoot+"expire.sql", now, uint16(10))
	})
	var state string
	if err := database.QueryRowContext(ctx, `SELECT lifecycle_state FROM `+schema+
		`.documents_knowledge_file_leases WHERE lease_id=$1`, upload).Scan(&state); err != nil ||
		state != "EXPIRED" {
		t.Fatalf("%s expired lease state=%s error=%v", schema, state, err)
	}
}
