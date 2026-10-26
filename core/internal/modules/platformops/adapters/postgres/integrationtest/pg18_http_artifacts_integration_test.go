//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func exerciseRecoveryHTTPArtifacts(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	now := time.Date(2026, 7, 23, 12, 5, 0, 0, time.UTC)
	source := "55555555-5555-4555-8555-555555555555"
	safety := "66666666-6666-4666-8666-666666666666"
	operationID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	sessionID := "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	query := `SELECT platform.platformops_request_restore_v2(
		$1,$2,$3,'existing_readable',$4,$5,$6,$7,NULL,NULL,false,$8,$9,$10,$11)`
	if _, err := database.ExecContext(ctx, query, operationID, source, environmentID,
		hashA, []byte(`{"restore":"http"}`), safety, hashB, now,
		primaryID, ownerID, sessionID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := database.QueryRowContext(ctx,
		`SELECT count(*) FROM platform.platformops_list_recovery_sets_v1($1)`,
		environmentID).Scan(&count); err != nil || count != 2 {
		t.Fatalf("recovery-set list count = %d / %v", count, err)
	}
	assertSingleUseArtifact(t, ctx, database,
		`SELECT * FROM platform.platformops_claim_recovery_manifest_v1($1,$2,$3,$4,$5)`,
		source, primaryID, ownerID, sessionID, now)
	wrongSession := "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
	var digest string
	var document []byte
	err := database.QueryRowContext(ctx,
		`SELECT * FROM platform.platformops_claim_restore_plan_v1($1,$2,$3,$4,$5)`,
		operationID, primaryID, ownerID, wrongSession, now.Add(time.Minute)).
		Scan(&digest, &document)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("wrong restore-plan session claim = %v", err)
	}
	assertSingleUseArtifact(t, ctx, database,
		`SELECT * FROM platform.platformops_claim_restore_plan_v1($1,$2,$3,$4,$5)`,
		operationID, primaryID, ownerID, sessionID, now.Add(time.Minute))
}

func assertSingleUseArtifact(
	t *testing.T, ctx context.Context, database *sql.DB,
	query string, arguments ...any,
) {
	t.Helper()
	var digest string
	var document []byte
	if err := database.QueryRowContext(ctx, query, arguments...).
		Scan(&digest, &document); err != nil || digest == "" || len(document) == 0 {
		t.Fatalf("first recovery artifact claim = %s/%d/%v", digest, len(document), err)
	}
	err := database.QueryRowContext(ctx, query, arguments...).Scan(&digest, &document)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("replayed recovery artifact claim = %v", err)
	}
}
