//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

const documentsPurgeSQLRoot = "adapters/postgres/purge/sql/tenant/"

func assertDocumentsPurgeSQL(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const (
		item        = "018f0000-0000-7000-8000-000000000202"
		version     = "018f0000-0000-7000-8000-000000000203"
		deletion    = "018f0000-0000-7000-8000-000000000501"
		operation   = "018f0000-0000-7000-8000-000000000502"
		audit       = "018f0000-0000-7000-8000-000000000503"
		finishOp    = "018f0000-0000-7000-8000-000000000504"
		finishAudit = "018f0000-0000-7000-8000-000000000505"
		lease       = "018f0000-0000-7000-8000-000000000506"
		bill        = "018f0000-0000-7000-8000-000000000507"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	tasks := []byte(`[{"deletion_id":"` + deletion + `","version_id":"` +
		version + `"}]`)
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		_, err := tx.ExecContext(ctx, `UPDATE documents_items
 SET lifecycle_state='TRASHED', trashed_at=$2, retention_ends_at=$3,
     row_version=row_version+1, updated_at=$2 WHERE id=$1`, item,
			now.Add(-31*24*time.Hour), now.Add(-time.Hour))
		if err != nil {
			t.Fatal(err)
		}
	})
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsPurgeSQLRoot+"begin.sql",
			item, now, tasks, operation, audit)
		var state string
		if err := tx.QueryRowContext(ctx, `SELECT lifecycle_state
 FROM documents_items WHERE id=$1`, item).Scan(&state); err != nil || state != "TRASHED" {
			t.Fatalf("blocked purge state = %s, error %v", state, err)
		}
	})
	releaseDocumentsReferences(t, ctx, database)
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsPurgeSQLRoot+"lock.sql", item, now)
		execDocumentsSQL(t, ctx, tx, documentsPurgeSQLRoot+"begin.sql",
			item, now, tasks, operation, audit)
	})
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/claim.sql", deletion, lease,
			now, now.Add(time.Minute))
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/begin_reconcile.sql",
			deletion, lease, int64(1))
		deleteHash := sha256.Sum256([]byte("version delete"))
		absenceHash := sha256.Sum256([]byte("version absent"))
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/confirm.sql", deletion, int64(1),
			"provider-version-delete", deleteHash[:], absenceHash[:], bill, now)
	})
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsPurgeSQLRoot+"finish.sql",
			item, now, finishOp, finishAudit)
	})
	var itemState, versionState string
	err := database.QueryRowContext(ctx, `SELECT i.lifecycle_state, v.lifecycle_state
 FROM documents_fresh_a.documents_items i
 JOIN documents_fresh_a.documents_versions v ON v.item_id=i.id
 WHERE i.id=$1`, item).Scan(&itemState, &versionState)
	if err != nil || itemState != "PURGED" || versionState != "PURGED" {
		t.Fatalf("purge states = %s/%s, error %v", itemState, versionState, err)
	}
}
