//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertAsyncConstraintCatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	assertAsyncForeignKeyPolicy(t, ctx, database)
	expected := map[string][]string{
		"jobs_pkey":                      {"PRIMARY KEY (id)"},
		"jobs_retention_check":           {"retain_until", "90 days", "365 days"},
		"jobs_parked_check":              {"parked_current_group_generation"},
		"job_attempts_job_fk":            {"scope_kind, scope_id, job_id"},
		"job_attempts_effect_receipt_fk": {"effect_receipt_id"},
		"outbox_stream_sequence_unique": {
			"scope_kind, scope_id, aggregate_type, aggregate_id, event_sequence",
		},
		"outbox_payload_check": {"payload_canonical", "sha256", "delivered"},
		"inbox_event_unique":   {"consumer_name, consumer_version, event_id"},
		"inbox_event_sequence_unique": {
			"consumer_name, consumer_version, scope_kind, scope_id",
			"aggregate_type, aggregate_id, event_sequence",
		},
		"idempotency_receipts_async_job_fk": {
			"scope_kind, scope_id, async_job_id",
			"platform.jobs(scope_kind, scope_id, id)",
		},
		"idempotency_receipts_effect_shape_check": {"effect_kind", "provider"},
		"dead_letters_root_fk":                    {"root_id"},
		"dead_letters_payload_hash_check":         {"payload_canonical", "sha256"},
	}
	for name, fragments := range expected {
		definition := asyncConstraintDefinition(t, ctx, database, name)
		for _, fragment := range fragments {
			if !strings.Contains(definition, fragment) {
				t.Fatalf("constraint %s lacks %q: %s", name, fragment, definition)
			}
		}
		if name == "outbox_stream_sequence_unique" &&
			strings.Contains(definition, "authority_") {
			t.Fatalf("outbox stream identity includes authority: %s", definition)
		}
	}
}

func assertAsyncForeignKeyPolicy(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	rows, err := database.QueryContext(ctx, `
SELECT c.conname, c.condeferrable, c.condeferred, c.confdeltype, c.confupdtype
FROM pg_catalog.pg_constraint c
JOIN pg_catalog.pg_class r ON r.oid = c.conrelid
JOIN pg_catalog.pg_namespace n ON n.oid = r.relnamespace
WHERE n.nspname = 'platform' AND c.contype = 'f'
  AND r.relname IN (
    'jobs', 'job_attempts', 'outbox', 'inbox',
    'idempotency_receipts', 'dead_letters'
  )`)
	if err != nil {
		t.Fatalf("read async foreign keys: %v", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var name, onDelete, onUpdate string
		var deferrable, deferred bool
		if err := rows.Scan(
			&name, &deferrable, &deferred, &onDelete, &onUpdate,
		); err != nil {
			t.Fatalf("scan async foreign key: %v", err)
		}
		if deferrable || deferred || onDelete != "r" || onUpdate != "a" {
			t.Fatalf("foreign key %s policy drifted", name)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate async foreign keys: %v", err)
	}
	if count != 28 {
		t.Fatalf("async foreign key count = %d, want 28", count)
	}
}

func asyncConstraintDefinition(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	name string,
) string {
	t.Helper()
	var definition string
	if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.pg_get_constraintdef(c.oid)
FROM pg_catalog.pg_constraint c
JOIN pg_catalog.pg_class r ON r.oid = c.conrelid
JOIN pg_catalog.pg_namespace n ON n.oid = r.relnamespace
WHERE n.nspname = 'platform' AND c.conname = $1`, name).Scan(&definition); err != nil {
		t.Fatalf("read constraint %s: %v", name, err)
	}
	return definition
}
