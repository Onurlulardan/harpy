//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const (
	documentsReconciliation   = "018f0000-0000-7000-8000-000000000080"
	documentsReverseRequester = "018f0000-0000-7000-8000-000000000081"
	documentsReverseApprover  = "018f0000-0000-7000-8000-000000000082"
	documentsReverseAccepted  = "2026-07-21T18:30:00+00:00"
)

func assertDocumentsProtectedReverse(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	withDocumentsPrincipal(t, ctx, database, "harpy_cutover_finalizer",
		"documents_adopted_a", func(connection *sql.Conn) {
			first, err := callDocumentsReverse(ctx, connection, 1)
			if err != nil || first == "" {
				t.Fatalf("reverse acceptance digest=%q err=%v", first, err)
			}
			replay, replayErr := callDocumentsReverse(ctx, connection, 1)
			if replayErr != nil || replay != first {
				t.Fatalf("reverse replay digest=%q err=%v", replay, replayErr)
			}
			_, changedErr := callDocumentsReverse(ctx, connection, 2)
			requireDocumentsPGError(t, changedErr, "22023", "55000")
		})
	withDocumentsPrincipal(t, ctx, database, "harpy_adopter",
		"documents_adopted_a", func(connection *sql.Conn) {
			_, err := callDocumentsReverse(ctx, connection, 1)
			requireDocumentsPGError(t, err, "42501")
		})
	_, err := database.ExecContext(ctx, `UPDATE
 documents_adopted_a.documents_v1_reverse_receipts
SET receipt_sha256=repeat('f',64)`)
	requireDocumentsPGError(t, err, "55000")
	assertDocumentsReverseCounts(t, ctx, database)
}

func callDocumentsReverse(ctx context.Context, connection *sql.Conn,
	targetCount int,
) (string, error) {
	const query = `SELECT documents_adopted_a.documents_v1_reverse_accept_v1(
$1::uuid,$2::uuid,12,13,1,$3::bigint,decode(repeat('e',64),'hex'),
decode(repeat('e',64),'hex'),$4::uuid,$5::uuid,$6::timestamptz)`
	var digest string
	err := connection.QueryRowContext(ctx, query,
		documentsReconciliation, documentsCutoverTenantID, targetCount,
		documentsReverseRequester, documentsReverseApprover,
		documentsReverseAccepted).Scan(&digest)
	return digest, err
}

func assertDocumentsReverseCounts(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var reconciliations, receipts int
	err := database.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM documents_adopted_a.documents_reverse_reconciliations),
 (SELECT count(*) FROM documents_adopted_a.documents_v1_reverse_receipts)`).Scan(
		&reconciliations, &receipts)
	if err != nil || reconciliations != 1 || receipts != 1 {
		t.Fatalf("reverse counts=%d/%d err=%v", reconciliations, receipts, err)
	}
}
