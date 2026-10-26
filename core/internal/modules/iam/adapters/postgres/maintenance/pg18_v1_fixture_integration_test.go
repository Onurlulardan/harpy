//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
)

const maintenancePG18V1Tables = `
CREATE TABLE organizations_form_table (
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,name varchar(255) NOT NULL,
 slug varchar(255) NOT NULL,status varchar(255),parent_id varchar(255));
CREATE TABLE roles_form_table (
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,name varchar(255) NOT NULL,
 organization_id varchar(255),is_default boolean);
CREATE TABLE users_form_table (
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,email varchar(255) NOT NULL,
 password varchar(255),role_ids text[],organization_id varchar(255),
 status varchar(255),token_version integer NOT NULL,account_type varchar(30) NOT NULL,
 portal_subject_type varchar(100),portal_subject_id uuid,
 portal_profile_form_slug varchar(255),portal_login_identifier varchar(100),
 portal_identifier_field varchar(255),portal_setup_required boolean NOT NULL,
 portal_setup_completed_at timestamp,workflow_status varchar(50));
CREATE TABLE refresh_tokens (
 id uuid PRIMARY KEY,user_id uuid NOT NULL,token_hash varchar(255) NOT NULL,
 expires_at timestamp NOT NULL);`

const maintenancePG18V1Rows = `
WITH organization_row AS (
 INSERT INTO organizations_form_table(id,tenant_id,name,slug,status,parent_id)
 VALUES('91000000-0000-4000-8000-000000000001',$1,
  'Legacy Organization','legacy-org','ACTIVE',NULL) RETURNING id
),role_row AS (
 INSERT INTO roles_form_table(id,tenant_id,name,organization_id,is_default)
 SELECT '92000000-0000-4000-8000-000000000001',$1,
  'Legacy Admin',id::text,true FROM organization_row RETURNING id
),user_row AS (
 INSERT INTO users_form_table(
  id,tenant_id,email,password,role_ids,organization_id,status,
  token_version,account_type,portal_setup_required)
 SELECT '93000000-0000-4000-8000-000000000001',$1,
  'Legacy.User@example.com','$2a$12$legacy-hash-evidence',ARRAY[id::text],
  '91000000-0000-4000-8000-000000000001','ACTIVE',7,'operator',false
 FROM role_row RETURNING id)
INSERT INTO refresh_tokens(id,user_id,token_hash,expires_at)
SELECT '94000000-0000-4000-8000-000000000001',id,
 'legacy-refresh-digest','2026-08-16T00:00:00Z' FROM user_row;`

func seedMaintenancePG18V1Snapshot(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, tenant := range maintenancePG18Tenants {
		execMaintenancePG18Tenant(
			t, ctx, database, tenant, maintenancePG18V1Tables,
		)
		execMaintenancePG18Tenant(
			t, ctx, database, tenant, maintenancePG18V1Rows, tenant.id,
		)
	}
}
