//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const createIAMV1Snapshot = `
CREATE TABLE organizations_form_table (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    name varchar(255) NOT NULL,
    slug varchar(255) NOT NULL,
    status varchar(255),
    parent_id varchar(255)
);
CREATE TABLE roles_form_table (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    name varchar(255) NOT NULL,
    organization_id varchar(255),
    is_default boolean
);
CREATE TABLE users_form_table (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    email varchar(255) NOT NULL,
    password varchar(255),
    role_ids text[],
    organization_id varchar(255),
    status varchar(255),
    token_version integer NOT NULL,
    account_type varchar(30) NOT NULL,
    portal_subject_type varchar(100),
    portal_subject_id uuid,
    portal_profile_form_slug varchar(255),
    portal_login_identifier varchar(100),
    portal_identifier_field varchar(255),
    portal_setup_required boolean NOT NULL,
    portal_setup_completed_at timestamp,
    workflow_status varchar(50)
);
CREATE TABLE refresh_tokens (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    token_hash varchar(255) NOT NULL,
    expires_at timestamp NOT NULL
);`

const insertIAMV1Snapshot = `
WITH organization_row AS (
    INSERT INTO organizations_form_table
        (id, tenant_id, name, slug, status, parent_id)
    VALUES ('91000000-0000-4000-8000-000000000001', $1,
        'Legacy Organization', 'legacy-org', 'ACTIVE', NULL)
    RETURNING id
), role_row AS (
    INSERT INTO roles_form_table
        (id, tenant_id, name, organization_id, is_default)
    SELECT '92000000-0000-4000-8000-000000000001', $1,
        'Legacy Admin', id::text, true FROM organization_row
    RETURNING id
), user_row AS (
    INSERT INTO users_form_table
        (id, tenant_id, email, password, role_ids,
         organization_id, status, token_version, account_type,
         portal_setup_required)
    SELECT '93000000-0000-4000-8000-000000000001', $1,
        'Legacy.User@example.com', '$2a$12$legacy-hash-evidence',
        ARRAY[id::text], '91000000-0000-4000-8000-000000000001',
        'ACTIVE', 7, 'operator', false FROM role_row
    RETURNING id
)
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
SELECT '94000000-0000-4000-8000-000000000001', id,
    'legacy-refresh-digest', '2026-08-16T00:00:00Z' FROM user_row;`

const insertIAMV1HybridSnapshot = `
INSERT INTO users_form_table (
    id, tenant_id, email, password, role_ids, organization_id,
    status, token_version, account_type, portal_subject_type,
    portal_subject_id, portal_profile_form_slug, portal_login_identifier,
    portal_identifier_field, portal_setup_required,
    portal_setup_completed_at, workflow_status
) VALUES (
    '93000000-0000-4000-8000-000000000002', $1,
    'hybrid.operator@example.com', '$2a$14$hybrid-hash-evidence',
    ARRAY[]::text[], NULL, 'ACTIVE', 4, 'operator_portal', 'operator',
    '93000000-0000-4000-8000-000000000002', 'users', 'OP-0002',
    NULL, false, NULL, 'DRAFT'
), (
    '93000000-0000-4000-8000-000000000003', $1,
    'rejected.hybrid@example.com', NULL,
    ARRAY[]::text[], NULL, 'INACTIVE', 2, 'operator_portal', 'operator',
    '93000000-0000-4000-8000-000000000003', 'users', 'OP-0003',
    NULL, true, NULL, 'REJECTED'
);`

func seedIAMV1AdoptedSnapshot(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(t, ctx, db, tenant.schema, createIAMV1Snapshot)
		mustExecIAMTenant(t, ctx, db, tenant.schema, insertIAMV1Snapshot, tenant.id)
		mustExecIAMTenant(t, ctx, db, tenant.schema, insertIAMV1HybridSnapshot, tenant.id)
	}
}
