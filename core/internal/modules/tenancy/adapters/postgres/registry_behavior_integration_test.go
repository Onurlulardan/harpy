//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertTenantRegistryCatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var tableColumns, viewColumns, utcTimestamps int
	err := database.QueryRowContext(ctx, `
SELECT
  count(*) FILTER (WHERE table_schema = 'platform' AND table_name = 'tenants'),
  count(*) FILTER (WHERE table_schema = 'public' AND table_name = 'tenants'),
  count(*) FILTER (
    WHERE table_schema = 'platform' AND table_name = 'tenants'
      AND column_name IN ('created_at', 'updated_at', 'deleted_at')
      AND data_type = 'timestamp with time zone'
  )
FROM information_schema.columns`).Scan(&tableColumns, &viewColumns, &utcTimestamps)
	if err != nil {
		t.Fatal(err)
	}
	if tableColumns != 14 || viewColumns != 9 || utcTimestamps != 3 {
		t.Fatalf("catalog columns table=%d view=%d utc=%d", tableColumns, viewColumns, utcTimestamps)
	}
}

func assertTenantRegistryWrites(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
UPDATE public.tenants SET
    name = 'Updated', status = 'TRIAL', settings = '{"updated":true}',
    owner_id = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',
    updated_at = TIMESTAMPTZ '2026-07-14 10:00:00+00',
    deleted_at = TIMESTAMPTZ '2026-07-14 11:00:00+00'
WHERE id = $1`, tenantID)
	if err != nil {
		t.Fatalf("V1 allowlisted update failed: %v", err)
	}
	var updated bool
	if err := database.QueryRowContext(ctx, `
SELECT name = 'Updated' AND status = 'TRIAL' AND settings = '{"updated":true}'::jsonb
FROM platform.tenants WHERE id = $1`, tenantID).Scan(&updated); err != nil || !updated {
		t.Fatalf("allowlisted update not persisted: value=%v error=%v", updated, err)
	}
	rejected := []string{
		"UPDATE public.tenants SET slug = slug || '_changed' WHERE id = $1",
		"UPDATE public.tenants SET created_at = statement_timestamp() WHERE id = $1",
		"UPDATE platform.tenants SET schema_name = 't_changed' WHERE id = $1",
		"DELETE FROM public.tenants WHERE id = $1",
		"DELETE FROM platform.tenants WHERE id = $1",
		`INSERT INTO public.tenants (id,name,slug,status,settings,owner_id)
		 VALUES ($1,'X','x','ACTIVE','{}',
		         'dddddddd-dddd-4ddd-8ddd-dddddddddddd')`,
	}
	for _, statement := range rejected {
		if _, err := database.ExecContext(ctx, statement, tenantID); err == nil ||
			!strings.Contains(err.Error(), "forbidden") &&
				!strings.Contains(err.Error(), "immutable") &&
				!strings.Contains(err.Error(), "denied") {
			t.Fatalf("mutation was not rejected by tenancy guard: %q error=%v", statement, err)
		}
	}
}

func assertAdoptionIssuesAreAppendOnly(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const issueID = "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee"
	const evidence = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_, err := database.ExecContext(ctx, `
INSERT INTO platform.tenant_adoption_issues (
    issue_id, tenant_id, issue_code, evidence_sha256
) VALUES ($1, $2, 'missing_schema', $3)`, issueID, freshTenantID, evidence)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"UPDATE platform.tenant_adoption_issues SET schema_name = 'tenant_x' WHERE issue_id = $1",
		"DELETE FROM platform.tenant_adoption_issues WHERE issue_id = $1",
	} {
		if _, err := database.ExecContext(ctx, statement, issueID); err == nil ||
			!strings.Contains(err.Error(), "append-only") {
			t.Fatalf("adoption issue mutation passed: %q error=%v", statement, err)
		}
	}
	_, err = database.ExecContext(ctx, `
INSERT INTO platform.tenant_adoption_issues (
    issue_id, issue_code, evidence_sha256
) VALUES ('ffffffff-ffff-4fff-8fff-ffffffffffff', 'extra_schema',
          'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb')`)
	if err == nil {
		t.Fatal("adoption issue without tenant or schema identity passed")
	}
}
