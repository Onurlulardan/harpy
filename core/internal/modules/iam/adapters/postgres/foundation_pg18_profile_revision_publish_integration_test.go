//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const createIAMProfileSchemaV2 = `
INSERT INTO iam_profile_schema_versions (
    id, resource_kind, version, state, definition_sha256,
    activated_at, retired_at, created_at, updated_at, revision
) VALUES ($1, 'user', 2, 'draft',
    pg_catalog.decode(pg_catalog.repeat('00', 32), 'hex'), NULL, NULL,
    '2026-07-16T01:30:00Z', '2026-07-16T01:30:00Z', 1);`

const retireIAMProfileField = `
UPDATE iam_profile_fields
SET state = 'retired', retired_at = '2026-07-16T01:40:00Z',
    updated_at = '2026-07-16T01:40:00Z', revision = revision + 1
WHERE id = $1;`

const insertIAMProfileSchemaV2Snapshot = `
INSERT INTO iam_profile_field_revisions (
    schema_version_id, resource_kind, field_id, validation_schema,
    validation_sha256, created_at, field_key, value_kind, cardinality,
    required, classification, normalization_profile, field_state,
    semantic_sha256
) SELECT $1, resource_kind, id, '{"maximum_bytes":96}'::jsonb,
    pg_catalog.sha256(pg_catalog.convert_to('{"maximum_bytes":96}', 'UTF8')),
    '2026-07-16T01:45:00Z', field_key, value_kind, cardinality, required,
    classification, normalization_profile, state,
    CASE WHEN $3::boolean THEN iam_profile_semantic_sha256_v1(
        id, resource_kind, field_key, value_kind, cardinality, required,
        classification, normalization_profile
    ) ELSE pg_catalog.decode(pg_catalog.repeat('ff', 32), 'hex') END
FROM iam_profile_fields WHERE id = $2;`

func publishIAMProfileSchemaV2(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	schema := iamFoundationTenants[0].schema
	mustExecIAMTenant(t, ctx, database, schema, createIAMProfileSchemaV2,
		iamProfileSchemaV2ID)
	mustExecIAMTenant(t, ctx, database, schema, retireIAMProfileField,
		iamProfileFieldID)

	invalidDigestErr := rejectIAMTenant(t, ctx, database, schema,
		insertIAMProfileSchemaV2Snapshot,
		iamProfileSchemaV2ID, iamProfileFieldID, false)
	assertIAMConstraint(t, invalidDigestErr, "23514",
		"iam_profile_field_revisions_semantic_sha256_check")

	mustExecIAMTenant(t, ctx, database, schema,
		insertIAMProfileSchemaV2Snapshot,
		iamProfileSchemaV2ID, iamProfileFieldID, true)
	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_profile_schema_versions
SET definition_sha256 = iam_profile_definition_sha256_v1(id),
    updated_at = '2026-07-16T01:50:00Z', revision = revision + 1
WHERE id = $1`, iamProfileSchemaV2ID)

	wrongHashErr := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_profile_schema_versions
SET state = 'active', activated_at = '2026-07-16T02:00:00Z',
    updated_at = '2026-07-16T02:00:00Z', revision = revision + 1,
    definition_sha256 = pg_catalog.decode(pg_catalog.repeat('ff', 32), 'hex')
WHERE id = $1`, iamProfileSchemaV2ID)
	assertIAMConstraint(t, wrongHashErr, "23514",
		"iam_profile_schema_versions_lifecycle_guard")

	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_profile_schema_versions
SET state = 'retired', retired_at = '2026-07-16T02:00:00Z',
    updated_at = '2026-07-16T02:00:00Z', revision = revision + 1
WHERE id = '41000000-0000-4000-8000-000000000001';
UPDATE iam_profile_schema_versions
SET state = 'active', activated_at = '2026-07-16T02:00:00Z',
    updated_at = '2026-07-16T02:00:00Z', revision = revision + 1
WHERE id = '44000000-0000-4000-8000-000000000010';`)
}
