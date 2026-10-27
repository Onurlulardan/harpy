//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
)

const maintenanceReadPermission = "93000000-0000-4000-8000-000000000001"

const maintenanceAuthoritySQL = `
INSERT INTO iam_resources(
 id,resource_key,owner_domain,resource_kind,display_name,description,
 is_protected,state,revision,created_at,updated_at,retired_at)
VALUES('91000000-0000-4000-8000-000000000001','iam.users','iam',
 'system_resource','Users',NULL,true,'active',1,
 '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL);
INSERT INTO iam_actions(
 id,action_key,display_name,description,is_protected,state,revision,
 created_at,updated_at,retired_at)
VALUES('92000000-0000-4000-8000-000000000001','read','Read',NULL,false,
 'active',1,'2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL);
INSERT INTO iam_permissions(
 id,match_kind,resource_id,action_id,is_protected,state,revision,
 created_at,updated_at,retired_at)
VALUES('93000000-0000-4000-8000-000000000001','exact',
 '91000000-0000-4000-8000-000000000001',
 '92000000-0000-4000-8000-000000000001',false,'active',1,
 '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL);
INSERT INTO iam_authorization_clock(
 singleton,clock_value,revision,created_at,updated_at)
VALUES(true,0,1,'2026-07-16T00:10:00Z','2026-07-16T00:10:00Z');
UPDATE iam_users SET authorization_version=1,revision=revision+1,
 updated_at='2026-07-16T00:20:00Z';
INSERT INTO iam_user_permission_bindings(
 id,user_id,permission_id,permission_match_kind,state,revision,
 created_at,updated_at,activated_at,revoked_at)
VALUES('a0040000-0000-4000-8000-000000000001',
 '51000000-0000-4000-8000-000000000003',
 '93000000-0000-4000-8000-000000000001','exact','active',1,
 '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',
 '2026-07-16T00:20:00Z',NULL);`

func seedMaintenancePG18Authority(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, tenant := range maintenancePG18Tenants {
		execMaintenancePG18Tenant(t, ctx, database, tenant, maintenanceAuthoritySQL)
	}
}

func prepareMaintenancePG18Runtime(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `
CREATE ROLE harpy_v2_runtime NOLOGIN;
GRANT harpy_runtime TO harpy_v2_runtime;
GRANT USAGE ON SCHEMA platform,t_iam_a,t_iam_b TO harpy_runtime;
ALTER TABLE platform.domain_cutovers
 ADD COLUMN state text NOT NULL DEFAULT 'v2',
 ADD COLUMN writer text NOT NULL DEFAULT 'v2';`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare maintenance runtime: %v", err)
	}
}
