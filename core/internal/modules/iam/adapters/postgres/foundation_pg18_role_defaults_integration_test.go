//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMDefaultRole = `
INSERT INTO iam_roles (
    id, scope_kind, scope_id, organization_id, name, name_key,
    description, status, is_protected, is_default, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, retired_at
) VALUES (
    $1, $2, $3, $4, $5, pg_catalog.convert_to($5, 'UTF8'),
    NULL, 'active', false, true, 'role',
    '41000000-0000-4000-8000-000000000003', '{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('43', 32), 'hex'),
    1, '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL
)`

func assertIAMDefaultRoleUniqueness(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	organizationID string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, insertIAMDefaultRole,
		"75000000-0000-4000-8000-000000000001", "organization",
		organizationID, organizationID, "Organization Default One")
	orgErr := rejectIAMTenant(t, ctx, database, schema, insertIAMDefaultRole,
		"75000000-0000-4000-8000-000000000002", "organization",
		organizationID, organizationID, "Organization Default Two")
	assertIAMConstraint(
		t, orgErr, "23505", "iam_roles_one_organization_default_uidx",
	)
	tenantID := iamFoundationTenants[0].id
	mustExecIAMTenant(t, ctx, database, schema, insertIAMDefaultRole,
		"75000000-0000-4000-8000-000000000003", "tenant",
		tenantID, nil, "Tenant Default One")
	tenantErr := rejectIAMTenant(t, ctx, database, schema, insertIAMDefaultRole,
		"75000000-0000-4000-8000-000000000004", "tenant",
		tenantID, nil, "Tenant Default Two")
	assertIAMConstraint(
		t, tenantErr, "23505", "iam_roles_one_tenant_default_uidx",
	)
}
