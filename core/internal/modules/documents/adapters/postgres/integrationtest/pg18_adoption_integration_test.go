//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

const documentsAdoptionSQLRoot = "adapters/postgres/adoption/sql/tenant/"

func assertDocumentsAdoptionSQL(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const (
		run           = "018f0000-0000-7000-8000-000000000701"
		sourceItem    = "018f0000-0000-7000-8000-000000000702"
		sourceVersion = "018f0000-0000-7000-8000-000000000703"
		targetItem    = "018f0000-0000-7000-8000-000000000704"
		targetVersion = "018f0000-0000-7000-8000-000000000705"
		lsn           = "0/16B6C50"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	manifest := sha256.Sum256([]byte("documents adoption manifest"))
	content := sha256.Sum256([]byte("legacy object bytes"))
	metadata := sha256.Sum256([]byte("canonical metadata"))
	cipher := sha256.Sum256([]byte("encrypted target object"))
	acceptance := sha256.Sum256([]byte("documents adoption acceptance"))
	targetKey := "tenants/ceirceircfardairceirceirce/documents/" + targetItem +
		"/versions/" + targetVersion + "/content"
	withDocumentsSchema(t, ctx, database, "documents_adopted_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsAdoptionSQLRoot+"create_run.sql",
			run, int64(17), manifest[:], "documents_adoption_v1", now)
		execDocumentsSQL(t, ctx, tx, documentsAdoptionSQLRoot+"inventory.sql",
			run, int64(17), manifest[:], sourceItem, sourceVersion,
			"legacy/documents/evidence.pdf", "etag-legacy-1", int64(19), content[:],
			metadata[:], lsn, targetItem, targetVersion, now)
		execDocumentsSQL(t, ctx, tx, documentsAdoptionSQLRoot+"prepare_copy.sql",
			run, int64(17), manifest[:], sourceVersion, now)
		execDocumentsSQL(t, ctx, tx, documentsAdoptionSQLRoot+"verify_copy.sql",
			run, int64(17), manifest[:], sourceVersion, "etag-legacy-1",
			targetVersion, targetKey, "target-object-version-1", int64(1), int64(129),
			content[:], cipher[:], "documents-kek-1", make([]byte, 115), metadata[:],
			lsn, now)
		execDocumentsSQL(t, ctx, tx, documentsAdoptionSQLRoot+"begin_shadow.sql",
			run, int64(17), manifest[:], now)
		execDocumentsSQL(t, ctx, tx, documentsAdoptionSQLRoot+"reconcile.sql",
			run, int64(17), manifest[:], lsn, lsn, int64(0), now)
		execDocumentsSQL(t, ctx, tx, documentsAdoptionSQLRoot+"accept.sql",
			run, int64(17), manifest[:], acceptance[:], lsn, now)
		execDocumentsSQL(t, ctx, tx, documentsAdoptionSQLRoot+"record_hit.sql",
			"legacy_sync_zip", "frontend-v2", "build-2026-07-18", now)
	})
	var state string
	var hits int
	err := database.QueryRowContext(ctx, `SELECT r.lifecycle_state,h.hit_count
FROM documents_adopted_a.documents_adoption_runs r
JOIN documents_adopted_a.documents_compatibility_hits h ON true
WHERE r.run_id=$1`, run).Scan(&state, &hits)
	if err != nil || state != "ACCEPTED" || hits != 1 {
		t.Fatalf("adoption state/hits = %s/%d, error %v", state, hits, err)
	}
}
