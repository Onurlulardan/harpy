//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const iamFieldClassificationAuthorityMigration = "20260730113200_provision_iam_field_classification_authority_v1"

func TestIAMFieldClassificationAuthorityAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(
				t, ctx, "field_classification_authority_"+origin,
			)
			prepareIAMFieldClassificationAuthority(t, ctx, database, origin)
			applyIAMFieldClassificationAuthority(t, ctx, database, origin)
			assertIAMFieldClassificationAuthority(t, ctx, database)
		})
	}
}

func prepareIAMFieldClassificationAuthority(
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
	resourceCatalog := iamFoundationDocument{
		path: "20260718030004_seed_iam_system_transfer_resources_v1." + origin + ".sql",
		content: readIAMIdentityMigration(
			t, origin, "20260718030004_seed_iam_system_transfer_resources_v1",
		),
	}
	for _, tenant := range iamFoundationTenants {
		seedIAMRBACCatalog(t, ctx, database, tenant.schema)
		seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
		applyIAMTenantDocuments(t, ctx, database, tenant.schema,
			[]iamFoundationDocument{permissionRead, resourceCatalog})
	}
	mustExecIAMTenant(
		t, ctx, database, iamFoundationTenants[0].schema,
		seedExistingIAMFieldClassificationAuthority,
	)
}

const seedExistingIAMFieldClassificationAuthority = `
INSERT INTO iam_actions (
    id, action_key, display_name, description, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES (
    '019fae00-0000-8000-8000-000000000901',
    'read.confidential', 'Adopted Confidential Read', NULL, false,
    'active', 1, '2026-07-30T00:00:00Z', '2026-07-30T00:00:00Z', NULL
);
INSERT INTO iam_permissions (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES (
    '019fae00-0000-8000-8000-000000000902', 'exact',
    '91000000-0000-4000-8000-000000000001',
    '019fae00-0000-8000-8000-000000000901', true,
    'active', 1, '2026-07-30T00:00:00Z', '2026-07-30T00:00:00Z', NULL
)`

func applyIAMFieldClassificationAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	document := iamFoundationDocument{
		path: iamFieldClassificationAuthorityMigration + "." + origin + ".sql",
		content: readIAMIdentityMigration(
			t, origin, iamFieldClassificationAuthorityMigration,
		),
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(
			t, ctx, database, tenant.schema, []iamFoundationDocument{document},
		)
	}
}

func assertIAMFieldClassificationAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for index, tenant := range iamFoundationTenants {
		connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
		var actions, permissions, tenantAllCoverage int
		var confidentialActionID, confidentialPermissionID string
		err := transaction.QueryRowContext(ctx, `
SELECT
    (SELECT count(*) FROM iam_actions
     WHERE action_key IN ('read.confidential','read.restricted')
       AND NOT is_protected AND state='active'),
    (SELECT count(*) FROM iam_permissions AS permission
     JOIN iam_resources AS resource ON resource.id=permission.resource_id
     JOIN iam_actions AS action ON action.id=permission.action_id
     WHERE resource.resource_key IN ('iam.users','iam.roles','iam.organizations')
       AND action.action_key IN ('read.confidential','read.restricted')
       AND permission.match_kind='exact' AND permission.is_protected
       AND permission.state='active'),
    (SELECT count(*) FROM iam_permissions AS permission
     JOIN iam_resources AS resource ON resource.id=permission.resource_id
     JOIN iam_actions AS action ON action.id=permission.action_id
     WHERE resource.resource_key IN ('iam.users','iam.roles','iam.organizations')
       AND action.action_key IN ('read.confidential','read.restricted')
       AND iam_user_has_permission_v1($1, permission.id)),
    (SELECT id::text FROM iam_actions WHERE action_key='read.confidential'),
    (SELECT permission.id::text FROM iam_permissions AS permission
     JOIN iam_resources AS resource ON resource.id=permission.resource_id
     JOIN iam_actions AS action ON action.id=permission.action_id
     WHERE resource.resource_key='iam.users'
       AND action.action_key='read.confidential')`,
			iamUserOneID,
		).Scan(
			&actions, &permissions, &tenantAllCoverage,
			&confidentialActionID, &confidentialPermissionID,
		)
		_ = transaction.Rollback()
		_ = connection.Close()
		if err != nil || actions != 2 || permissions != 6 || tenantAllCoverage != 6 {
			t.Fatalf("tenant %s classification authority=%d/%d/%d: %v",
				tenant.schema, actions, permissions, tenantAllCoverage, err)
		}
		if index == 0 &&
			(confidentialActionID != "019fae00-0000-8000-8000-000000000901" ||
				confidentialPermissionID != "019fae00-0000-8000-8000-000000000902") {
			t.Fatalf("adopted semantic identities changed: %s/%s",
				confidentialActionID, confidentialPermissionID)
		}
	}
}
