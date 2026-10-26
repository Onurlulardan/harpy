//go:build integration

package postgres

const insertIAMOrganization = `
INSERT INTO iam_organizations (
    id, name, slug, status, parent_id, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, tombstoned_at
) VALUES (
    $1, $2, $3, 'active', $4, 'organization',
    '41000000-0000-4000-8000-000000000002', '{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('41', 32), 'hex'),
    1, '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL
)`

const insertIAMMembership = `
INSERT INTO iam_organization_memberships (
    id, organization_id, user_id, membership_kind, state,
    revision, created_at, updated_at, tombstoned_at
) VALUES (
    $1, $2, $3, 'member', 'active', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL
)`

const insertIAMOrganizationRole = `
INSERT INTO iam_roles (
    id, scope_kind, scope_id, organization_id, name, name_key,
    description, status, is_protected, is_default, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, retired_at
) VALUES (
    $1, 'organization', $2, $2, $3, pg_catalog.convert_to($3, 'UTF8'),
    NULL, 'active', false, false, 'role',
    '41000000-0000-4000-8000-000000000003', '{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('42', 32), 'hex'),
    1, '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL
)`

const insertIAMOrganizationRoleAssignment = `
INSERT INTO iam_user_role_assignments (
    id, user_id, role_id, role_scope_kind, role_scope_id,
    organization_id, membership_id, state, revision,
    created_at, updated_at, revoked_at
) VALUES (
    $1, $2, $3, 'organization', $4,
    $4, $5, 'active', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL
)`
