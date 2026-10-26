//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	iamProfileFieldID    = "43000000-0000-4000-8000-000000000010"
	iamProfileSchemaV2ID = "44000000-0000-4000-8000-000000000010"
)

const insertIAMProfileFieldFixture = `
INSERT INTO iam_profile_fields (
    id, resource_kind, field_key, value_kind, cardinality, required,
    classification, normalization_profile, state, created_at, updated_at,
    retired_at, revision
) VALUES ($1, 'user', 'display_name', 'string', 'single', false,
    'internal', 'none_v1', 'active',
    '2026-07-16T01:00:00Z', '2026-07-16T01:00:00Z', NULL, 1);`

const insertIAMProfileSnapshotFixture = `
INSERT INTO iam_profile_field_revisions (
    schema_version_id, resource_kind, field_id, validation_schema,
    validation_sha256, created_at
) VALUES ($2, 'user', $1,
    pg_catalog.jsonb_build_object('maximum_bytes', $3::integer),
    pg_catalog.sha256(pg_catalog.convert_to(
        '{"maximum_bytes":' || $3::text || '}', 'UTF8'
    )), '2026-07-16T01:00:00Z');`

func seedIAMProfileRevisionFixtures(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for index, tenant := range iamFoundationTenants {
		maximumBytes := 64 + index*64
		mustExecIAMTenant(t, ctx, database, tenant.schema,
			insertIAMProfileFieldFixture, iamProfileFieldID)
		mustExecIAMTenant(t, ctx, database, tenant.schema,
			insertIAMProfileSnapshotFixture,
			iamProfileFieldID, iamUserProfileSchemaID, maximumBytes)
	}
}

func applyIAMProfileRevisionMigrations(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamProfileRevisionMigrationNames))
	for _, name := range iamProfileRevisionMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}
