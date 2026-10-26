//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestDocumentsCutoverPG18ReplayAuthorityAndReverse(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	database := startDocumentsPG18(t, ctx)
	prepareDocumentsPG18(t, ctx, database)
	prepareDocumentsCutoverPlatform(t, ctx, database)
	applyDocumentsMigrations(t, ctx, database)
	prepareDocumentsCutoverIdentities(t, ctx, database, documentsCutoverTenantID)
	seedDocumentsCutoverOperation(t, ctx, database, "documents_adopted_a")

	withDocumentsPrincipal(t, ctx, database, "harpy_adopter",
		"documents_adopted_a", func(connection *sql.Conn) {
			applyDocumentsInitialCutover(t, ctx, connection)
			assertDocumentsAdopterDenials(t, ctx, connection)
		})
	assertDocumentsCutoverCounts(t, ctx, database, 1, 0)
	assertDocumentsFreshOriginDenied(t, ctx, database)

	completeDocumentsCopyEffect(t, ctx, database)
	withDocumentsPrincipal(t, ctx, database, "harpy_adopter",
		"documents_adopted_a", func(connection *sql.Conn) {
			keys := []string{"inventory", "folder", "item", "version"}
			for index, key := range keys {
				body := documentsCutoverBatch(t, key, "complete", 2,
					string(rune('5'+index)), []any{})
				if _, err := callDocumentsComposite(ctx, connection,
					"documents_adopted_a", key, "complete", 2,
					string(rune('5'+index)), body); err != nil {
					t.Fatalf("complete %s: %v", key, err)
				}
			}
		})
	assertDocumentsCutoverCounts(t, ctx, database, 1, 1)
	assertDocumentsReceiptImmutable(t, ctx, database)

	switchDocumentsAuthorityToV2(t, ctx, database)
	assertDocumentsProtectedReverse(t, ctx, database)
}

func applyDocumentsInitialCutover(t *testing.T, ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	keys := []string{"inventory", "folder", "item", "version"}
	for index, key := range keys {
		body := documentsCutoverBatch(t, key, "delta", 1,
			string(rune('1'+index)), documentsCutoverChanges(key, "Evidence"))
		result, err := callDocumentsComposite(ctx, connection,
			"documents_adopted_a", key, "delta", 1,
			string(rune('1'+index)), body)
		if err != nil || result.stateSHA256 == "" {
			t.Fatalf("apply %s: result=%+v err=%v", key, result, err)
		}
		if key == "inventory" {
			replay, replayErr := callDocumentsComposite(ctx, connection,
				"documents_adopted_a", key, "delta", 1, "1", body)
			if replayErr != nil || replay != result {
				t.Fatalf("inventory replay=%+v err=%v", replay, replayErr)
			}
		}
	}
	tampered := documentsCutoverBatch(t, "folder", "delta", 1, "2",
		documentsCutoverChanges("folder", "Changed"))
	_, err := callDocumentsComposite(ctx, connection, "documents_adopted_a",
		"folder", "delta", 1, "2", tampered)
	requireDocumentsPGError(t, err, "22000")
}

func assertDocumentsAdopterDenials(t *testing.T, ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	_, err := connection.ExecContext(ctx, `INSERT INTO documents_reverse_reconciliations (
 reconciliation_id,source_generation,target_generation,source_count,target_count,
 source_sha256,target_sha256,lifecycle_state,requested_by,created_at)
VALUES ('018f0000-0000-7000-8000-000000000070',12,13,0,0,
 decode(repeat('e',64),'hex'),decode(repeat('e',64),'hex'),'PENDING',
 '018f0000-0000-7000-8000-000000000071',clock_timestamp())`)
	requireDocumentsPGError(t, err, "42501")
	_, err = connection.ExecContext(ctx, `SELECT
 documents_v1_adoption_mapping_ready_v1($1,$2)`,
		documentsCutoverOperation, documentsCutoverRun)
	requireDocumentsPGError(t, err, "42501")
	_, err = connection.ExecContext(ctx, `SELECT
 platform.platform_async_enqueue_documents_cutover_copy_v1(
 $1,$2,$3,$4,$5,clock_timestamp())`, documentsCutoverOperation,
		documentsCutoverRun, documentsAdoptionRun, documentsSourceVersion,
		documentsCopyJob)
	requireDocumentsPGError(t, err, "42501")
}

func assertDocumentsFreshOriginDenied(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	withDocumentsPrincipal(t, ctx, database, "harpy_adopter",
		"documents_fresh_a", func(connection *sql.Conn) {
			body := documentsCutoverBatch(t, "inventory", "delta", 3, "9",
				documentsCutoverChanges("inventory", "Evidence"))
			_, err := callDocumentsComposite(ctx, connection, "documents_fresh_a",
				"inventory", "delta", 3, "9", body)
			requireDocumentsPGError(t, err, "22023")
		})
}
