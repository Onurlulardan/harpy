//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMProfileRevisionGuards(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	schema := iamFoundationTenants[0].schema
	semanticErr := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_profile_fields
SET field_key = 'rewritten_name', revision = revision + 1,
    updated_at = '2026-07-16T01:10:00Z'
WHERE id = '43000000-0000-4000-8000-000000000010'`)
	assertIAMConstraint(t, semanticErr, "23514", "iam_profile_fields_lifecycle_guard")

	snapshotErr := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_profile_field_revisions
SET validation_schema = '{"maximum_bytes":32}'::jsonb
WHERE schema_version_id = '41000000-0000-4000-8000-000000000001'
  AND field_id = '43000000-0000-4000-8000-000000000010'`)
	assertIAMSQLState(t, snapshotErr, "55000")

	for name, statement := range map[string]string{
		"schema": `DELETE FROM iam_profile_schema_versions
            WHERE id = '41000000-0000-4000-8000-000000000001'`,
		"field": `DELETE FROM iam_profile_fields
            WHERE id = '43000000-0000-4000-8000-000000000010'`,
		"snapshot": `DELETE FROM iam_profile_field_revisions
            WHERE schema_version_id =
                  '41000000-0000-4000-8000-000000000001'`,
		"truncate": "TRUNCATE TABLE iam_profile_field_revisions",
	} {
		t.Run(name, func(t *testing.T) {
			assertIAMSQLState(t,
				rejectIAMTenant(t, ctx, database, schema, statement), "55000")
		})
	}
}
