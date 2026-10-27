//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"slices"
	"testing"
)

func assertAsyncSliceACatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	assertAsyncTableCatalog(t, ctx, database)
	assertAsyncColumnNullability(t, ctx, database)
	assertAsyncConstraintCatalog(t, ctx, database)
}

func assertAsyncTableCatalog(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	rows, err := database.QueryContext(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'platform'
  AND table_name IN (
    'jobs', 'job_attempts', 'outbox', 'inbox',
    'idempotency_receipts', 'dead_letters'
  )
ORDER BY table_name`)
	if err != nil {
		t.Fatalf("read async table catalog: %v", err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatalf("scan async table catalog: %v", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate async table catalog: %v", err)
	}
	want := []string{
		"dead_letters", "idempotency_receipts", "inbox",
		"job_attempts", "jobs", "outbox",
	}
	if !slices.Equal(tables, want) {
		t.Fatalf("async tables = %v, want %v", tables, want)
	}
}

func assertAsyncColumnNullability(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	expected := map[string]bool{
		"jobs.id": false, "jobs.tenant_id": true,
		"job_attempts.attempt_id": false, "job_attempts.effect_receipt_id": true,
		"outbox.envelope_contract_version":       false,
		"outbox.business_effect_idempotency_key": false,
		"outbox.payload_canonical":               true, "outbox.payload_schema_sha256": false,
		"inbox.consumer_version": false, "inbox.causation_id": true,
		"idempotency_receipts.effect_kind":  true,
		"idempotency_receipts.provider":     true,
		"idempotency_receipts.async_job_id": true,
		"dead_letters.payload_canonical":    true,
		"dead_letters.redacted_envelope":    false,
	}
	for identity, nullable := range expected {
		var actual string
		if err := database.QueryRowContext(ctx, `
SELECT is_nullable
FROM information_schema.columns
WHERE table_schema = 'platform'
  AND table_name || '.' || column_name = $1`, identity).Scan(&actual); err != nil {
			t.Fatalf("read nullability for %s: %v", identity, err)
		}
		if (actual == "YES") != nullable {
			t.Fatalf("%s nullable = %s, want %t", identity, actual, nullable)
		}
	}
}
