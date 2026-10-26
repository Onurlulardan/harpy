//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertIAMRestQueryTenant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	marker string,
) {
	t.Helper()
	actual := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT profile_values ->> $1
FROM iam_rest_query_users_v1
WHERE id='51000000-0000-4000-8000-000000000001'`, restQueryTenantFieldID)
	if actual != marker {
		t.Fatalf("%s marker = %q, want %q", schema, actual, marker)
	}
	integer := mustRunIAMAtomicBigint(t, ctx, database, schema, `
SELECT (profile_values ->> $1)::bigint
FROM iam_rest_query_users_v1
WHERE id='51000000-0000-4000-8000-000000000001'`, restQueryIntegerFieldID)
	if integer != 42 {
		t.Fatalf("%s typed custom integer = %d", schema, integer)
	}
	missing := mustRunIAMAtomicBigint(t, ctx, database, schema, `
SELECT pg_catalog.count(*)
FROM iam_rest_query_users_v1
WHERE NOT (profile_values ? $1)
  AND (profile_values ->> $1) IS NULL`, restQueryIntegerFieldID)
	if missing != 1 {
		t.Fatalf("%s missing custom null count = %d", schema, missing)
	}
	displayName := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT display_name
FROM iam_rest_query_users_v2
WHERE id='51000000-0000-4000-8000-000000000001'`)
	if displayName != "Straße "+marker {
		t.Fatalf("%s display name = %q", schema, displayName)
	}
	legacyDisplayName := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT display_name
FROM iam_rest_query_users_v2
WHERE id='51000000-0000-4000-8000-000000000002'`)
	if legacyDisplayName != "Legacy "+marker {
		t.Fatalf("%s legacy display name = %q", schema, legacyDisplayName)
	}
	userMatches := mustRunIAMAtomicBigint(t, ctx, database, schema, `
SELECT pg_catalog.count(*)
FROM iam_rest_query_users_v2
WHERE display_name_folded COLLATE pg_catalog.pg_unicode_fast
      LIKE '%strasse%'`)
	roleMatches := mustRunIAMAtomicBigint(t, ctx, database, schema, `
SELECT pg_catalog.count(*)
FROM iam_rest_query_roles_v2
WHERE name_folded COLLATE pg_catalog.pg_unicode_fast LIKE '%strasse%'
   OR description_folded COLLATE pg_catalog.pg_unicode_fast LIKE '%strasse%'`)
	organizationMatches := mustRunIAMAtomicBigint(t, ctx, database, schema, `
SELECT pg_catalog.count(*)
FROM iam_rest_query_organizations_v2
WHERE name_folded COLLATE pg_catalog.pg_unicode_fast LIKE '%münchen%'
   OR slug_folded COLLATE pg_catalog.pg_unicode_fast LIKE '%münchen%'`)
	if userMatches != 1 || roleMatches != 1 || organizationMatches != 1 {
		t.Fatalf("%s search matches = %d/%d/%d, want 1/1/1",
			schema, userMatches, roleMatches, organizationMatches)
	}
	ownerProjection := mustRunIAMAtomicBigint(t, ctx, database, schema, `
SELECT pg_catalog.count(*)
FROM iam_rest_query_organizations_v3
WHERE owner_ids = ARRAY[
        '51000000-0000-4000-8000-000000000001'::uuid
      ]
  AND user_ids @> ARRAY[
        '51000000-0000-4000-8000-000000000001'::uuid,
        '51000000-0000-4000-8000-000000000002'::uuid
      ]`)
	if ownerProjection != 1 {
		t.Fatalf("%s organization owner projection = %d, want 1",
			schema, ownerProjection)
	}
	versions := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT token_version::text || ':' || authorization_version::text
FROM iam_rest_query_actor_versions_v1(
    pg_catalog.current_schema()::name,
    '51000000-0000-4000-8000-000000000001'
)`)
	if versions != "1:1" {
		t.Fatalf("%s actor versions = %q", schema, versions)
	}
	assertIAMRestQueryIndexPlan(t, ctx, database, schema, marker)
}

func assertIAMRestQueryIndexPlan(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	marker string,
) {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = transaction.Rollback()
		_ = connection.Close()
	}()
	if _, err := transaction.ExecContext(ctx, "SET LOCAL enable_seqscan=off"); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	rows, err := transaction.QueryContext(ctx, `
EXPLAIN (COSTS OFF)
SELECT id FROM iam_rest_query_users_v1
WHERE profile_values @> pg_catalog.jsonb_build_object($1::text,$2::text)`,
		restQueryTenantFieldID, marker)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		plan.WriteString(line)
		plan.WriteByte('\n')
	}
	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plan.String(), "iam_users_profile_values_gin_idx") {
		t.Fatalf("%s plan does not use custom GIN index:\n%s", schema, plan.String())
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
