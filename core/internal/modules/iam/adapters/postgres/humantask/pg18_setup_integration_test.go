//go:build integration

package humantask

import (
	"context"
	"database/sql"
	"testing"
)

const iamHumanTenantA = "76000000-0000-4000-8000-000000000071"
const iamHumanTenantB = "76000000-0000-4000-8000-000000000072"

func prepareIAMHumanPG18(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	const base = `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime NOLOGIN;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE TABLE platform.cutover_groups(scope_kind text,scope_id uuid,group_id text,
 generation bigint,state text,writer text,PRIMARY KEY(scope_kind,scope_id,group_id));
CREATE TABLE platform.cutover_group_members(scope_kind text,scope_id uuid,group_id text,
 ordinal smallint,domain text,PRIMARY KEY(scope_kind,scope_id,group_id,ordinal));
CREATE TABLE platform.domain_cutovers(scope_kind text,scope_id uuid,domain text,
 generation bigint,state text,writer text,PRIMARY KEY(scope_kind,scope_id,domain));
INSERT INTO platform.cutover_groups VALUES
 ('tenant','` + iamHumanTenantA + `','iam_system_resources',7,'v2','v2'),
 ('tenant','` + iamHumanTenantB + `','iam_system_resources',9,'v2','v2');
INSERT INTO platform.cutover_group_members VALUES
 ('tenant','` + iamHumanTenantA + `','iam_system_resources',1,'iam'),
 ('tenant','` + iamHumanTenantA + `','iam_system_resources',2,'system_resource_layout'),
 ('tenant','` + iamHumanTenantB + `','iam_system_resources',1,'iam'),
 ('tenant','` + iamHumanTenantB + `','iam_system_resources',2,'system_resource_layout');
INSERT INTO platform.domain_cutovers VALUES
 ('tenant','` + iamHumanTenantA + `','iam',7,'v2','v2'),
 ('tenant','` + iamHumanTenantA + `','system_resource_layout',7,'v2','v2'),
 ('tenant','` + iamHumanTenantB + `','iam',9,'v2','v2'),
 ('tenant','` + iamHumanTenantB + `','system_resource_layout',9,'v2','v2');`
	if _, err := database.ExecContext(ctx, base); err != nil {
		t.Fatal(err)
	}
	for index, schema := range []string{"iam_human_a", "iam_human_b"} {
		tenant := []string{iamHumanTenantA, iamHumanTenantB}[index]
		if err := createIAMHumanSchema(t, ctx, database, schema, tenant, origin); err != nil {
			t.Fatal(err)
		}
	}
	applyIAMHumanMigrations(t, ctx, database, origin)
}

func createIAMHumanSchema(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema, tenant, origin string,
) error {
	t.Helper()
	_, err := database.ExecContext(ctx, `SELECT pg_catalog.set_config('search_path','pg_catalog',false);
CREATE SCHEMA `+schema+` AUTHORIZATION harpy_owner;
CREATE TABLE `+schema+`.tenancy_schema_identity(singleton boolean PRIMARY KEY,
 tenant_id uuid NOT NULL,schema_name name NOT NULL);
INSERT INTO `+schema+`.tenancy_schema_identity VALUES(true,'`+tenant+`','`+schema+`');
CREATE TABLE `+schema+`.iam_resources(id uuid PRIMARY KEY,resource_key text UNIQUE,
 owner_domain text,resource_kind text,display_name text,description text,
 is_protected boolean,state text,revision bigint,created_at timestamptz,
 updated_at timestamptz,retired_at timestamptz);
CREATE TABLE `+schema+`.iam_actions(id uuid PRIMARY KEY,action_key text UNIQUE,
 display_name text,description text,is_protected boolean,state text,revision bigint,
 created_at timestamptz,updated_at timestamptz,retired_at timestamptz);
CREATE TABLE `+schema+`.iam_permissions(id uuid PRIMARY KEY,match_kind text,
 resource_id uuid,action_id uuid,is_protected boolean,state text,revision bigint,
 created_at timestamptz,updated_at timestamptz,retired_at timestamptz,
 UNIQUE(match_kind,resource_id,action_id));`)
	if err != nil || origin != "v1_adopted" {
		return err
	}
	_, err = database.ExecContext(ctx, `INSERT INTO `+schema+`.iam_resources VALUES(
'10000000-0000-4000-8000-000000000001','workflow.human_tasks','workflow',
'human_task','Adopted',NULL,false,'active',4,now(),now(),NULL)`)
	return err
}
