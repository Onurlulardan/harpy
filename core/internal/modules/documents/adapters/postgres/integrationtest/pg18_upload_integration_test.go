//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const documentsUploadSQLRoot = "adapters/postgres/upload/sql/tenant/"

func assertDocumentsUploadSQL(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const (
		uploadID   = "018f0000-0000-7000-8000-000000000201"
		document   = "018f0000-0000-7000-8000-000000000202"
		version    = "018f0000-0000-7000-8000-000000000203"
		operation  = "018f0000-0000-7000-8000-000000000204"
		audit      = "018f0000-0000-7000-8000-000000000205"
		reserve    = "018f0000-0000-7000-8000-000000000206"
		lease      = "018f0000-0000-7000-8000-000000000207"
		receipt    = "018f0000-0000-7000-8000-000000000208"
		readyOp    = "018f0000-0000-7000-8000-000000000209"
		readyAudit = "018f0000-0000-7000-8000-000000000210"
		actor      = "22222222-2222-4222-8222-222222222222"
		folder     = "018f0000-0000-7000-8000-000000000101"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	requestHash := sha256.Sum256([]byte("upload request"))
	contentHash := sha256.Sum256([]byte("content"))
	cipherHash := sha256.Sum256([]byte("ciphertext"))
	objectKey := "tenants/ceirceircfardairceirceirce/documents/" + document +
		"/versions/" + version + "/content"
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"create.sql",
			uploadID, actor, document, version, nil, folder, requestHash[:], contentHash[:],
			int64(7), int64(117), "evidence.pdf", "application/pdf", "proof", `["legal"]`,
			reserve, now.Add(time.Hour), now, operation, audit, "upload-sql-1")
		assertDocumentsTxCount(t, ctx, tx, "documents_uploads", 1)
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"begin.sql",
			uploadID, actor, lease, receipt)
		assertDocumentsTxCount(t, ctx, tx, "documents_upload_attempts", 1)
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"record_effect.sql",
			uploadID, actor, int64(1), objectKey, "opaque-version-1", int64(1), int64(7),
			int64(117), contentHash[:], cipherHash[:], "application/pdf", "pdf",
			"documents-kek-1", make([]byte, 115), "clamav-20260718", now)
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"prepare_finalize.sql",
			uploadID, actor, int64(1), "opaque-version-1")
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"finalize.sql",
			uploadID, actor, int64(1), objectKey, "opaque-version-1", int64(1), int64(7),
			int64(117), contentHash[:], cipherHash[:], "application/pdf", "pdf",
			"documents-kek-1", make([]byte, 115), "clamav-20260718", now, readyOp, readyAudit)
	})
	var uploadState, itemState, versionState string
	err := database.QueryRowContext(ctx, `SELECT u.lifecycle_state, i.lifecycle_state,
 v.lifecycle_state FROM documents_fresh_a.documents_uploads u
 JOIN documents_fresh_a.documents_items i ON i.id = u.document_id
 JOIN documents_fresh_a.documents_versions v ON v.id = u.candidate_version_id
 WHERE u.id = $1`, uploadID).Scan(&uploadState, &itemState, &versionState)
	if err != nil || uploadState != "READY" || itemState != "ACTIVE" || versionState != "READY" {
		t.Fatalf("upload/item/version states = %s/%s/%s, error %v",
			uploadState, itemState, versionState, err)
	}
	assertDocumentsOrphanReconcileSQL(t, ctx, database, actor, folder)
}

func assertDocumentsOrphanReconcileSQL(t *testing.T, ctx context.Context,
	database *sql.DB, actor, folder string,
) {
	t.Helper()
	const (
		uploadID  = "018f0000-0000-7000-8000-000000000221"
		document  = "018f0000-0000-7000-8000-000000000222"
		version   = "018f0000-0000-7000-8000-000000000223"
		operation = "018f0000-0000-7000-8000-000000000224"
		audit     = "018f0000-0000-7000-8000-000000000225"
		reserve   = "018f0000-0000-7000-8000-000000000226"
		lease     = "018f0000-0000-7000-8000-000000000227"
		receipt   = "018f0000-0000-7000-8000-000000000228"
		deletion  = "018f0000-0000-7000-8000-000000000229"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	requestHash := sha256.Sum256([]byte("orphan request"))
	contentHash := sha256.Sum256([]byte("orphan content"))
	objectKey := "tenants/ceirceircfardairceirceirce/documents/" + document +
		"/versions/" + version + "/content"
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"create.sql",
			uploadID, actor, document, version, nil, folder, requestHash[:], contentHash[:],
			int64(7), int64(117), "orphan.pdf", "application/pdf", "", `[]`,
			reserve, now.Add(time.Hour), now, operation, audit, "upload-orphan-1")
		execDocumentsSQL(t, ctx, tx, documentsUploadSQLRoot+"begin.sql",
			uploadID, actor, lease, receipt)
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/uploadreconcile/sql/tenant/prepare.sql",
			uploadID, int64(1))
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/uploadreconcile/sql/tenant/finalize.sql",
			uploadID, int64(1), objectKey, "opaque-orphan-version", int64(7),
			int64(117), deletion, now.Add(10*time.Minute))
	})
	var uploadState, attemptState, deletionState string
	err := database.QueryRowContext(ctx, `SELECT u.lifecycle_state, a.lifecycle_state,
 d.lifecycle_state FROM documents_fresh_a.documents_uploads u
 JOIN documents_fresh_a.documents_upload_attempts a ON a.upload_id = u.id
 JOIN documents_fresh_a.documents_blob_deletions d ON d.target_id = u.id
 WHERE u.id = $1`, uploadID).Scan(&uploadState, &attemptState, &deletionState)
	if err != nil || uploadState != "RETRYABLE" || attemptState != "CLEANUP_PENDING" ||
		deletionState != "PENDING" {
		t.Fatalf("orphan states=%s/%s/%s err=%v", uploadState, attemptState,
			deletionState, err)
	}
}

func assertDocumentsTxCount(t *testing.T, ctx context.Context, tx *sql.Tx,
	table string, expected int,
) {
	t.Helper()
	var actual int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&actual); err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("%s count = %d, want %d", table, actual, expected)
	}
}

func withDocumentsSchema(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, action func(*sql.Tx),
) {
	t.Helper()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path', $1, true)",
		schema+", pg_catalog"); err != nil {
		t.Fatal(err)
	}
	action(tx)
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func execDocumentsSQL(t *testing.T, ctx context.Context, tx *sql.Tx,
	path string, arguments ...any,
) {
	t.Helper()
	path = filepath.Join("..", "..", "..", filepath.FromSlash(path))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(ctx, string(body), arguments...); err != nil {
		t.Fatalf("execute %s: %v", path, err)
	}
}
