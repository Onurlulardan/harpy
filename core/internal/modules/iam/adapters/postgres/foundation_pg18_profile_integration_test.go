//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertSecondActiveIAMProfileSchema = `
INSERT INTO iam_profile_schema_versions (
    id, resource_kind, version, state, definition_sha256,
    activated_at, retired_at, created_at, updated_at, revision
) VALUES (
    $1, $2, $3, 'active',
    pg_catalog.decode(pg_catalog.repeat('33', 32), 'hex'),
    '2026-07-16T00:00:00Z', NULL,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', 1
)`

func assertIAMProfileConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	schema := iamFoundationTenants[0].schema
	activeErr := rejectIAMTenant(t, ctx, database, schema,
		insertSecondActiveIAMProfileSchema,
		"42000000-0000-4000-8000-000000000001", "user", 2)
	assertIAMConstraint(
		t,
		activeErr,
		"23505",
		"iam_profile_schema_versions_one_active_resource_idx",
	)

	fieldID := "43000000-0000-4000-8000-000000000001"
	mustExecIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_profile_fields (
    id, resource_kind, field_key, value_kind, cardinality, required,
    classification, normalization_profile, state, created_at, updated_at,
    retired_at, revision
) VALUES ($1, 'user', 'display_name', 'string', 'single', false,
    'internal', 'none_v1', 'active',
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL, 1)`, fieldID)
	immutableErr := rejectIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_profile_fields (
    id, resource_kind, field_key, value_kind, cardinality, required,
    classification, normalization_profile, state, created_at, updated_at,
    retired_at, revision
) VALUES ($1, 'user', 'changed_semantics', 'integer', 'list', true,
    'restricted', 'none_v1', 'active',
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL, 1)`, fieldID)
	assertIAMConstraint(t, immutableErr, "23505", "iam_profile_fields_pkey")
	draftSchemaID := "44000000-0000-4000-8000-000000000001"
	mustExecIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_profile_schema_versions (
    id, resource_kind, version, state, definition_sha256,
    activated_at, retired_at, created_at, updated_at, revision
) VALUES ($1, 'user', 2, 'draft',
    pg_catalog.decode(pg_catalog.repeat('45', 32), 'hex'),
    NULL, NULL, '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', 1)`,
		draftSchemaID)
	mustExecIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_profile_field_revisions (
    schema_version_id, resource_kind, field_id, validation_schema,
    validation_sha256, created_at
) VALUES ($1, 'user', $2, '{"maxLength": 80}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('46', 32), 'hex'),
    '2026-07-16T00:00:00Z')`, draftSchemaID, fieldID)

	mismatchErr := rejectIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_profile_field_revisions (
    schema_version_id, resource_kind, field_id, validation_schema,
    validation_sha256, created_at
) VALUES ($1, 'organization', $2, '{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('44', 32), 'hex'),
    '2026-07-16T00:00:00Z')`, iamUserProfileSchemaID, fieldID)
	assertIAMSQLState(t, mismatchErr, "23503")

	mustExecIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_profile_field_revisions (
    schema_version_id, resource_kind, field_id, validation_schema,
    validation_sha256, created_at
) VALUES ($1, 'user', $2, '{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('44', 32), 'hex'),
    '2026-07-16T00:00:00Z')`, iamUserProfileSchemaID, fieldID)
}
