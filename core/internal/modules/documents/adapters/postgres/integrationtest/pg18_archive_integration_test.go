//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

const documentsArchiveSQLRoot = "adapters/postgres/archive/sql/tenant/"

func assertDocumentsArchiveSQL(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const (
		archiveID = "018f0000-0000-7000-8000-000000000401"
		jobID     = "018f0000-0000-7000-8000-000000000402"
		reserve   = "018f0000-0000-7000-8000-000000000403"
		operation = "018f0000-0000-7000-8000-000000000404"
		audit     = "018f0000-0000-7000-8000-000000000405"
		deletion  = "018f0000-0000-7000-8000-000000000406"
		lease     = "018f0000-0000-7000-8000-000000000407"
		bill      = "018f0000-0000-7000-8000-000000000408"
		actor     = "22222222-2222-4222-8222-222222222222"
		document  = "018f0000-0000-7000-8000-000000000202"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	request := []byte(`{"item_ids":["` + document + `"]}`)
	requestHash := sha256.Sum256(request)
	artifactHash := sha256.Sum256([]byte("zip artifact"))
	cipherHash := sha256.Sum256([]byte("encrypted zip artifact"))
	objectKey := "tenants/ceirceircfardairceirceirce/document-archives/" +
		archiveID + "/content"
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsArchiveSQLRoot+"create.sql",
			archiveID, actor, int64(9), jobID, nil, requestHash[:], request,
			reserve, now, operation, "archive-sql-1", audit)
		execDocumentsSQL(t, ctx, tx, documentsArchiveSQLRoot+"pin.sql", archiveID, now)
		assertDocumentsTxCount(t, ctx, tx, "documents_archive_entries", 1)
		execDocumentsSQL(t, ctx, tx, documentsArchiveSQLRoot+"record_effect.sql",
			archiveID, int64(256), int64(366), artifactHash[:], cipherHash[:],
			objectKey, "opaque-archive-version", int64(1), "documents-kek-1",
			make([]byte, 115), now)
		execDocumentsSQL(t, ctx, tx, documentsArchiveSQLRoot+"mark_ready.sql",
			archiveID, int64(256), int64(366), artifactHash[:], cipherHash[:],
			objectKey, "opaque-archive-version", int64(1), "documents-kek-1",
			make([]byte, 115), now)
		execDocumentsSQL(t, ctx, tx, documentsArchiveSQLRoot+"expire.sql",
			archiveID, deletion, now.Add(24*time.Hour))
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/claim.sql", deletion, lease,
			now.Add(24*time.Hour), now.Add(24*time.Hour+time.Minute))
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/begin_reconcile.sql",
			deletion, lease, int64(1))
		deleteHash := sha256.Sum256([]byte("delete receipt"))
		absenceHash := sha256.Sum256([]byte("absence receipt"))
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/confirm.sql",
			deletion, int64(1), "provider-request-1", deleteHash[:],
			absenceHash[:], bill, now.Add(24*time.Hour))
	})
	var state string
	var count int
	err := database.QueryRowContext(ctx, `SELECT lifecycle_state, entry_count
 FROM documents_fresh_a.documents_archives WHERE id = $1`, archiveID).
		Scan(&state, &count)
	if err != nil || state != "PURGED" || count != 1 {
		t.Fatalf("archive state/count = %s/%d, error %v", state, count, err)
	}
}
