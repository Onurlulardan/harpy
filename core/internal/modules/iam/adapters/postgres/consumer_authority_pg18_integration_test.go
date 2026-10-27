//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

var iamConsumerAuthorityMigrations = []string{
	"20260716014004_create_iam_rbac_user_bumps_v1",
	"20260718161200_reconcile_analytics_authority_v1",
	"20260802234000_publish_iam_analytics_dashboard_export_v1",
	"20260802234001_publish_iam_support_bug_report_authority_v1",
	"20260802234002_grant_iam_support_bug_report_authority_v1",
}

func TestIAMConsumerAuthoritiesAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "consumer_authority_"+origin)
			prepareIAMConsumerAuthority(t, ctx, database, origin)
			applyIAMConsumerAuthority(t, ctx, database, origin)
			assertIAMConsumerAuthority(t, ctx, database)
		})
	}
}

func prepareIAMConsumerAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	applyIAMFoundation(t, ctx, database, origin)
	seedIAMFoundationBase(t, ctx, database)
	applyIAMLifecycleGuards(t, ctx, database, origin)
	applyIAMRBAC(t, ctx, database, origin)
	for _, tenant := range iamFoundationTenants {
		seedIAMRBACCatalog(t, ctx, database, tenant.schema)
		seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
		mustExecIAMTenant(t, ctx, database, tenant.schema, seedIAMSupportRoles, tenant.id)
	}
}

const seedIAMSupportRoles = `
INSERT INTO iam_roles (
    id,scope_kind,scope_id,organization_id,name,name_key,description,status,
    is_protected,is_default,profile_resource_kind,profile_schema_version_id,
    profile_values,profile_sha256,revision,created_at,updated_at,retired_at
) VALUES
('97000000-0000-4000-8000-000000000001','tenant',$1,NULL,
 'Temel Kullanıcı',pg_catalog.convert_to('Temel Kullanıcı','UTF8'),NULL,
 'active',false,true,'role','41000000-0000-4000-8000-000000000003',
 '{}'::jsonb,pg_catalog.decode(pg_catalog.repeat('92',32),'hex'),1,
 '2026-08-02T00:00:00Z','2026-08-02T00:00:00Z',NULL),
('97000000-0000-4000-8000-000000000002','tenant',$1,NULL,
 'Tenant Yöneticisi',pg_catalog.convert_to('Tenant Yöneticisi','UTF8'),NULL,
 'active',false,false,'role','41000000-0000-4000-8000-000000000003',
 '{}'::jsonb,pg_catalog.decode(pg_catalog.repeat('93',32),'hex'),1,
 '2026-08-02T00:00:00Z','2026-08-02T00:00:00Z',NULL)
`

func applyIAMConsumerAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamConsumerAuthorityMigrations))
	for _, name := range iamConsumerAuthorityMigrations {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func assertIAMConsumerAuthority(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
		var analyticsExport, supportPermissions, reporterGrants int
		var baseGrants, adminGrants int
		err := transaction.QueryRowContext(ctx, `
SELECT
 (SELECT pg_catalog.count(*) FROM iam_permissions AS permission
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  JOIN iam_actions AS action ON action.id=permission.action_id
  WHERE resource.resource_key='analytics.dashboard'
    AND action.action_key='export' AND permission.state='active'),
 (SELECT pg_catalog.count(*) FROM iam_permissions AS permission
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  WHERE resource.resource_key='bug_reports' AND permission.state='active'),
 (SELECT pg_catalog.count(*) FROM iam_role_permission_bindings AS binding
  JOIN iam_permissions AS permission ON permission.id=binding.permission_id
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  WHERE resource.resource_key='bug_reports' AND binding.state='active'),
 (SELECT pg_catalog.count(*) FROM iam_role_permission_bindings AS binding
  JOIN iam_permissions AS permission ON permission.id=binding.permission_id
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  WHERE resource.resource_key='bug_reports' AND binding.state='active'
    AND binding.role_id='97000000-0000-4000-8000-000000000001'),
 (SELECT pg_catalog.count(*) FROM iam_role_permission_bindings AS binding
  JOIN iam_permissions AS permission ON permission.id=binding.permission_id
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  WHERE resource.resource_key='bug_reports' AND binding.state='active'
    AND binding.role_id='97000000-0000-4000-8000-000000000002')`).Scan(
			&analyticsExport, &supportPermissions, &reporterGrants, &baseGrants, &adminGrants,
		)
		_ = transaction.Rollback()
		_ = connection.Close()
		if err != nil || analyticsExport != 1 || supportPermissions != 4 ||
			reporterGrants != 10 || baseGrants != 2 || adminGrants != 4 {
			t.Fatalf("tenant %s consumer authority=%d/%d/%d/%d/%d: %v",
				tenant.schema, analyticsExport, supportPermissions, reporterGrants,
				baseGrants, adminGrants, err)
		}
	}
}
