//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const seedIAMFormsRecordAuthorityReuse = `
ALTER TABLE iam_permission_implications
    DROP CONSTRAINT iam_permission_implications_kind_check;
ALTER TABLE iam_permission_implications
    ADD CONSTRAINT iam_permission_implications_kind_check CHECK (
        implication_kind IN ('v1_form_parent_v1', 'forms_record_operation_v1')
    );
INSERT INTO iam_resources (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    'e6000000-0000-4000-8000-000000000001', 'forms.records', 'forms',
    'system_resource', 'Existing Forms Records', NULL, true, 'active', 1,
    '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL
);
INSERT INTO iam_actions (
    id, action_key, display_name, description, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES
    ('e6000000-0000-4000-8000-000000000101', 'create', 'Existing Create',
     NULL, false, 'active', 1,
     '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL),
    ('e6000000-0000-4000-8000-000000000105', 'import', 'Existing Import',
     NULL, false, 'active', 1,
     '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL);
INSERT INTO iam_permissions (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES
    ('e6000000-0000-4000-8000-000000000201', 'exact',
     'e6000000-0000-4000-8000-000000000001',
     'e6000000-0000-4000-8000-000000000101', false, 'active', 1,
     '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL),
    ('e6000000-0000-4000-8000-000000000202', 'exact',
     'e6000000-0000-4000-8000-000000000001',
     '92000000-0000-4000-8000-000000000001', false, 'active', 1,
     '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL),
    ('e6000000-0000-4000-8000-000000000205', 'exact',
     'e6000000-0000-4000-8000-000000000001',
     'e6000000-0000-4000-8000-000000000105', false, 'active', 1,
     '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL);
INSERT INTO iam_permission_implications (
    id, implication_kind, source_permission_id, source_match_kind,
    implied_permission_id, implied_match_kind, retirement_contract_key,
    state, hit_count, last_hit_at, retirement_receipt_id,
    revision, created_at, updated_at, retired_at
) VALUES (
    'e6000000-0000-4000-8000-000000000301', 'forms_record_operation_v1',
    'e6000000-0000-4000-8000-000000000205', 'exact',
    'e6000000-0000-4000-8000-000000000201', 'exact',
    'forms.records.operation.v1', 'active', 0, NULL, NULL, 1,
    '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL
)`

const seedIAMFormsRecordAuthorityDrift = `
INSERT INTO iam_resources (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    'e6000000-0000-4000-8000-000000000901', 'forms.records', 'iam',
    'system_resource', 'Drifted Forms Records', NULL, true, 'active', 1,
    '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL
)`

func prepareIAMFormsRecordAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	applyIAMFoundation(t, ctx, database, origin)
	seedIAMFoundationBase(t, ctx, database)
	applyIAMLifecycleGuards(t, ctx, database, origin)
	applyIAMRBAC(t, ctx, database, origin)
	permissionRead := iamFoundationDocument{
		path: "20260716013433_create_iam_user_has_permission_v1." + origin + ".sql",
		content: readIAMIdentityMigration(
			t, origin, "20260716013433_create_iam_user_has_permission_v1",
		),
	}
	for _, tenant := range iamFoundationTenants {
		seedIAMRBACCatalog(t, ctx, database, tenant.schema)
		seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
		applyIAMTenantDocuments(t, ctx, database, tenant.schema,
			[]iamFoundationDocument{permissionRead})
	}
	mustExecIAMTenant(t, ctx, database, iamFoundationTenants[0].schema,
		seedIAMFormsRecordAuthorityReuse)
	mustExecIAMTenant(t, ctx, database, iamFoundationTenants[1].schema,
		seedIAMFormsRecordAuthorityDrift)
}

func applyIAMFormsRecordAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	document := iamFoundationDocument{
		path: iamFormsRecordAuthorityMigration + "." + origin + ".sql",
		content: readIAMIdentityMigration(
			t, origin, iamFormsRecordAuthorityMigration,
		),
	}
	applyIAMTenantDocuments(t, ctx, database, iamFoundationTenants[0].schema,
		[]iamFoundationDocument{document})
	err := rejectIAMTenant(t, ctx, database, iamFoundationTenants[1].schema,
		string(document.content))
	assertIAMConstraint(t, err, "23514", "iam_resources_revision_guard")
}
