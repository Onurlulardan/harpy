//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

type documentsApplyResult struct {
	count       int64
	keySHA256   string
	stateSHA256 string
}

func withDocumentsPrincipal(t *testing.T, ctx context.Context, database *sql.DB,
	principal, schema string, action func(*sql.Conn),
) {
	t.Helper()
	allowed := map[string]bool{
		"harpy_adopter": true, "harpy_cutover_finalizer": true,
	}
	allowedSchema := map[string]bool{
		"documents_adopted_a": true, "documents_adopted_b": true,
		"documents_fresh_a": true,
	}
	if !allowed[principal] || !allowedSchema[schema] {
		t.Fatal("invalid principal fixture")
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	_, err = connection.ExecContext(ctx, "SET SESSION AUTHORIZATION "+principal+
		"; SET search_path TO "+schema+",pg_catalog;")
	if err != nil {
		t.Fatal(err)
	}
	defer connection.ExecContext(context.Background(), "RESET SESSION AUTHORIZATION")
	action(connection)
}

func callDocumentsComposite(ctx context.Context, connection *sql.Conn, schema,
	checkpointKey, state string, ordinal int, sourceXID string, body []byte,
) (documentsApplyResult, error) {
	digest := sha256.Sum256(body)
	commitLSN := "0/20"
	var finalLSN any
	lastKey := fmt.Sprintf(`{"source":%q}`, checkpointKey)
	if state == "complete" {
		commitLSN, finalLSN, lastKey = "0/30", "0/30", "null"
	}
	query := `SELECT applied_count,key_state_sha256,state_sha256 FROM ` + schema +
		`.documents_v1_adoption_apply_and_record_v1(
$1::uuid,$2::uuid,$3::text,$4::uuid,$5::smallint,$6::text,$7::text,
$8::pg_lsn,$9::pg_lsn,$10::pg_lsn,$11::pg_lsn,$12::xid8,$13::integer,
$14::text,$15::jsonb,$16::bytea,$17::text,$18::bigint,$19::text)`
	var result documentsApplyResult
	err := connection.QueryRowContext(ctx, query,
		documentsCutoverOperation, documentsCutoverRun, "tenant",
		documentsCutoverTenantID, 1, checkpointKey, "documents_snapshot",
		"0/10", commitLSN, commitLSN, finalLSN, sourceXID, ordinal,
		state, lastKey, body, hex.EncodeToString(digest[:]), len(body),
		documentsCutoverManifest).Scan(
		&result.count, &result.keySHA256, &result.stateSHA256)
	return result, err
}

func requireDocumentsPGError(t *testing.T, err error, codes ...string) {
	t.Helper()
	var postgres *pgconn.PgError
	if !errors.As(err, &postgres) {
		t.Fatalf("expected PostgreSQL error, got %v", err)
	}
	for _, code := range codes {
		if postgres.Code == code {
			return
		}
	}
	t.Fatalf("PostgreSQL code=%s message=%s want=%v", postgres.Code,
		postgres.Message, codes)
}

func assertDocumentsCutoverCounts(t *testing.T, ctx context.Context,
	database *sql.DB, wantJobs, wantFolders int,
) {
	t.Helper()
	var jobs, bindings, checkpoints, folders, items, versions, receipts, tenantB int
	err := database.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM platform.jobs),
 (SELECT count(*) FROM platform.cutover_job_bindings),
 (SELECT count(*) FROM documents_adopted_a.documents_blob_copy_checkpoints),
 (SELECT count(*) FROM documents_adopted_a.documents_folders),
 (SELECT count(*) FROM documents_adopted_a.documents_items),
 (SELECT count(*) FROM documents_adopted_a.documents_versions),
 (SELECT count(*) FROM documents_adopted_a.documents_v1_cutover_receipts),
 (SELECT count(*) FROM documents_adopted_b.documents_v1_adoption_stage)`).Scan(
		&jobs, &bindings, &checkpoints, &folders, &items, &versions, &receipts, &tenantB)
	if err != nil || jobs != wantJobs || bindings != wantJobs || checkpoints != wantJobs ||
		folders != wantFolders || items != wantFolders || versions != wantFolders ||
		receipts != wantFolders || tenantB != 0 {
		t.Fatalf("cutover counts=%d/%d/%d/%d/%d/%d/%d tenantB=%d err=%v",
			jobs, bindings, checkpoints, folders, items, versions, receipts, tenantB, err)
	}
}
