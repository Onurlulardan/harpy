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
	documentsLinkSQLRoot = "adapters/postgres/link/sql/tenant/"
	documentsHoldSQLRoot = "adapters/postgres/legalhold/sql/tenant/"
	documentsLinkID      = "018f0000-0000-7000-8000-000000000601"
	documentsHoldID      = "018f0000-0000-7000-8000-000000000604"
)

func assertDocumentsReferenceSQL(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const (
		actor         = "22222222-2222-4222-8222-222222222222"
		item          = "018f0000-0000-7000-8000-000000000202"
		source        = "018f0000-0000-7000-8000-000000000611"
		sourceVersion = "018f0000-0000-7000-8000-000000000612"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	linkHash := sha256.Sum256([]byte("document lifecycle link"))
	holdHash := sha256.Sum256([]byte("document legal hold"))
	evidence := sha256.Sum256([]byte("case evidence"))
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsLinkSQLRoot+"register.sql",
			documentsLinkID, actor, item, "FORMS", "FORM_RECORD", source,
			sourceVersion, "LIFECYCLE", "018f0000-0000-7000-8000-000000000602",
			int64(1), now, linkHash[:], "link-sql-1",
			"018f0000-0000-7000-8000-000000000603")
		execDocumentsSQL(t, ctx, tx, documentsHoldSQLRoot+"place.sql",
			documentsHoldID, actor, item, "FORMS", "CASE-2026-1", evidence[:], now,
			"018f0000-0000-7000-8000-000000000605", "hold-sql-1", holdHash[:],
			"018f0000-0000-7000-8000-000000000606")
		assertDocumentsTxCount(t, ctx, tx, "documents_links", 1)
		assertDocumentsTxCount(t, ctx, tx, "documents_legal_holds", 1)
	})
}

func releaseDocumentsReferences(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const (
		actor         = "22222222-2222-4222-8222-222222222222"
		source        = "018f0000-0000-7000-8000-000000000611"
		sourceVersion = "018f0000-0000-7000-8000-000000000612"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	linkHash := sha256.Sum256([]byte("release document lifecycle link"))
	holdHash := sha256.Sum256([]byte("release document legal hold"))
	withDocumentsSchema(t, ctx, database, "documents_fresh_a", func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx, documentsLinkSQLRoot+"release.sql",
			documentsLinkID, actor, "FORMS", "FORM_RECORD", source, sourceVersion, now,
			"018f0000-0000-7000-8000-000000000607", "link-release-sql-1",
			linkHash[:], "018f0000-0000-7000-8000-000000000608")
		execDocumentsSQL(t, ctx, tx, documentsHoldSQLRoot+"release.sql",
			documentsHoldID, actor, "FORMS", "CASE-2026-1", now,
			"018f0000-0000-7000-8000-000000000609", "hold-release-sql-1",
			holdHash[:], "018f0000-0000-7000-8000-000000000610")
		if _, err := tx.ExecContext(ctx, "SAVEPOINT immutable_reference"); err != nil {
			t.Fatal(err)
		}
		_, err := tx.ExecContext(ctx, `UPDATE documents_links
 SET lifecycle_state='ACTIVE', released_at=NULL WHERE id=$1`, documentsLinkID)
		if err == nil {
			t.Fatal("released document reference was reactivated")
		}
		if _, err = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT immutable_reference"); err != nil {
			t.Fatal(err)
		}
	})
}
