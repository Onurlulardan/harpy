//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func applyIAMPlatformDocuments(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	documents []iamFoundationDocument,
) {
	t.Helper()
	for _, document := range documents {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin %s: %v", document.path, err)
		}
		if _, err := transaction.ExecContext(ctx, string(document.content)); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("apply %s: %v", document.path, err)
		}
		if err := transaction.Commit(); err != nil {
			t.Fatalf("commit %s: %v", document.path, err)
		}
	}
}

func applyIAMTenantDocuments(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	documents []iamFoundationDocument,
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve %s migration connection: %v", schema, err)
	}
	defer connection.Close()
	for _, document := range documents {
		transaction, err := connection.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin %s for %s: %v", document.path, schema, err)
		}
		var bound string
		if err := transaction.QueryRowContext(
			ctx,
			"SELECT pg_catalog.set_config('search_path', $1, true)",
			schema+", pg_catalog",
		).Scan(&bound); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("bind %s for %s: %v", schema, document.path, err)
		}
		if bound != schema+", pg_catalog" {
			_ = transaction.Rollback()
			t.Fatalf("bound search_path = %q, want %q", bound, schema+", pg_catalog")
		}
		if _, err := transaction.ExecContext(ctx, string(document.content)); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("apply %s to %s: %v", document.path, schema, err)
		}
		if err := transaction.Commit(); err != nil {
			t.Fatalf("commit %s for %s: %v", document.path, schema, err)
		}
	}
}
