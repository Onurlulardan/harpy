//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

func assertIAMAuthorizationFence(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var actorDocument string
	err = transaction.QueryRowContext(ctx, `
SELECT actor_document::text FROM iam_authorization_evidence_read_v2(
    pg_catalog.current_schema()::name, $1::uuid,
    '51000000-0000-4000-8000-000000000001', 'iam.users', 'read'
)`, tenant.id).Scan(&actorDocument)
	if err != nil || !strings.Contains(actorDocument, "51000000") {
		_ = transaction.Rollback()
		t.Fatalf("establish IAM authorization fence: %v / %s", err, actorDocument)
	}
	blockedContext, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	var blockedActor string
	err = database.QueryRowContext(blockedContext, `
SELECT id::text FROM `+tenant.schema+`.iam_users
WHERE id='51000000-0000-4000-8000-000000000001' FOR UPDATE`).Scan(&blockedActor)
	if err == nil {
		_ = transaction.Rollback()
		t.Fatal("authorization actor fence allowed concurrent revocation lock")
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit IAM authorization fence: %v", err)
	}
	releasedContext, releasedCancel := context.WithTimeout(ctx, 2*time.Second)
	defer releasedCancel()
	if err := database.QueryRowContext(releasedContext, `
SELECT id::text FROM `+tenant.schema+`.iam_users
WHERE id='51000000-0000-4000-8000-000000000001' FOR UPDATE`).Scan(&blockedActor); err != nil {
		t.Fatalf("authorization actor fence was not released: %v", err)
	}
}

func assertIAMAuthorizationSnapshot(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	tx, err := connection.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err = tx.ExecContext(ctx, statement); err != nil {
			_ = tx.Rollback()
			t.Fatal(err)
		}
	}
	var bound string
	err = tx.QueryRowContext(ctx,
		"SELECT pg_catalog.set_config('search_path', $1, true)",
		tenant.schema+", pg_catalog",
	).Scan(&bound)
	if err != nil || bound != tenant.schema+", pg_catalog" {
		_ = tx.Rollback()
		t.Fatalf("bind read-only authorization snapshot: %v / %s", err, bound)
	}
	var actorDocument string
	err = tx.QueryRowContext(ctx, `
SELECT actor_document::text FROM iam_authorization_evidence_read_v3(
    pg_catalog.current_schema()::name, $1::uuid,
    '51000000-0000-4000-8000-000000000001', 'iam.users', 'read'
)`, tenant.id).Scan(&actorDocument)
	if err != nil || !strings.Contains(actorDocument, "51000000") {
		_ = tx.Rollback()
		t.Fatalf("read-only V3 authorization evidence: %v / %s", err, actorDocument)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(fmt.Errorf("commit authorization snapshot: %w", err))
	}
}
