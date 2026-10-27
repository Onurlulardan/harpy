//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

func seedDocumentsRecoveryUpload(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture, now time.Time,
	request, content [32]byte,
) {
	t.Helper()
	withDocumentsSchema(t, ctx, database, fixture.schema, func(tx *sql.Tx) {
		_, err := tx.ExecContext(ctx, `INSERT INTO documents_uploads(
 id,actor_id,document_id,candidate_version_id,folder_id,request_sha256,
 expected_content_sha256,expected_plaintext_bytes,ciphertext_upper_bound,
 original_name,declared_mime,billing_reservation_id,attempt_generation,
 lifecycle_state,cleanup_required,expires_at,created_at,updated_at)
 VALUES ('018f0000-0000-7000-8000-000000000a30',$1,
 '018f0000-0000-7000-8000-000000000a31',
 '018f0000-0000-7000-8000-000000000a32',$2,$3,$4,7,117,
 'pending.pdf','application/pdf',
 '018f0000-0000-7000-8000-000000000a33',0,'CREATED',false,
 $5,$6,$6)`, documentsRecoveryOwner, documentsRecoveryFolder,
			request[:], content[:], now.Add(time.Hour), now)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func seedDocumentsRecoveryArchives(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture, now time.Time,
) {
	t.Helper()
	pending := sha256.Sum256([]byte("recovery pending archive"))
	ready := sha256.Sum256([]byte("recovery ready archive"))
	entries := sha256.Sum256([]byte("recovery archive entries"))
	artifact := sha256.Sum256([]byte("recovery archive artifact"))
	cipher := sha256.Sum256([]byte("recovery archive ciphertext"))
	withDocumentsSchema(t, ctx, database, fixture.schema, func(tx *sql.Tx) {
		_, err := tx.ExecContext(ctx, `INSERT INTO documents_archives(
 id,actor_id,authorization_revision,job_id,request_sha256,request_document,
 billing_reservation_id,lifecycle_state,entry_count,plaintext_bytes,
 created_at,updated_at)
 VALUES ('018f0000-0000-7000-8000-000000000a40',$1,1,
 '018f0000-0000-7000-8000-000000000a41',$2,$3,
 '018f0000-0000-7000-8000-000000000a42','PENDING',0,0,$4,$4)`,
			documentsRecoveryOwner, pending[:], []byte("{}"), now)
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO documents_archives(
 id,actor_id,authorization_revision,job_id,request_sha256,request_document,
 billing_reservation_id,lifecycle_state,entry_count,entries_sha256,
 plaintext_bytes,ciphertext_bytes,artifact_sha256,artifact_plaintext_bytes,
 ciphertext_sha256,cipher_format,object_key,object_version,object_generation,
 key_id,wrapped_dek,ready_at,expires_at,created_at,updated_at)
 VALUES ('018f0000-0000-7000-8000-000000000a43',$1,1,
 '018f0000-0000-7000-8000-000000000a44',$2,$3,
 '018f0000-0000-7000-8000-000000000a45','READY',1,$4,7,117,
 $5,7,$6,'FCDOC1',$7,'archive-source-version',1,'documents-kek-1',
 $8,$9,$10,$9,$9)`,
				documentsRecoveryOwner, ready[:], []byte("{}"), entries[:],
				artifact[:], cipher[:], "tenants/"+fixture.tenantID+
					"/document-archives/recovery/content", make([]byte, 115),
				now, now.Add(24*time.Hour))
		}
		if err != nil {
			t.Fatal(err)
		}
	})
}
