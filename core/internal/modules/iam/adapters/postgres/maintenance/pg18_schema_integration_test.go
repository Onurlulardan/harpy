//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
)

type maintenancePG18Tenant struct {
	id     string
	schema string
}

var maintenancePG18Tenants = []maintenancePG18Tenant{
	{"10000000-0000-4000-8000-000000000001", "t_iam_a"},
	{"20000000-0000-4000-8000-000000000002", "t_iam_b"},
}

const maintenancePG18SchemaSQL = `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_adopter LOGIN NOINHERIT;
CREATE SCHEMA platform;
CREATE TABLE platform.tenants (id uuid PRIMARY KEY,schema_name name NOT NULL UNIQUE);
CREATE SCHEMA t_iam_a;
CREATE SCHEMA t_iam_b;
INSERT INTO platform.tenants (id,schema_name) VALUES
 ('10000000-0000-4000-8000-000000000001','t_iam_a'),
 ('20000000-0000-4000-8000-000000000002','t_iam_b');
CREATE TABLE platform.environment_identity (environment_id uuid PRIMARY KEY);
CREATE TABLE platform.domain_cutovers (
 scope_kind text NOT NULL,scope_id uuid NOT NULL,domain text NOT NULL,
 generation bigint NOT NULL,PRIMARY KEY(scope_kind,scope_id,domain),
 UNIQUE(scope_kind,scope_id,domain,generation));
CREATE TABLE platform.cutover_groups (
 scope_kind text NOT NULL,scope_id uuid NOT NULL,group_id text NOT NULL,
 generation bigint NOT NULL,PRIMARY KEY(scope_kind,scope_id,group_id),
 UNIQUE(scope_kind,scope_id,group_id,generation));
CREATE TABLE platform.cutover_group_members (
 scope_kind text NOT NULL,scope_id uuid NOT NULL,group_id text NOT NULL,
 ordinal smallint NOT NULL,domain text NOT NULL,
 PRIMARY KEY(scope_kind,scope_id,group_id,ordinal),
 FOREIGN KEY(scope_kind,scope_id,group_id)
  REFERENCES platform.cutover_groups(scope_kind,scope_id,group_id),
 FOREIGN KEY(scope_kind,scope_id,domain)
  REFERENCES platform.domain_cutovers(scope_kind,scope_id,domain),
 UNIQUE(scope_kind,scope_id,group_id,domain));
INSERT INTO platform.environment_identity VALUES
 ('30000000-0000-4000-8000-000000000001');
INSERT INTO platform.domain_cutovers VALUES
 ('tenant','10000000-0000-4000-8000-000000000001','iam',1),
 ('tenant','10000000-0000-4000-8000-000000000001','system_resource_layout',1),
 ('tenant','20000000-0000-4000-8000-000000000002','iam',1),
 ('tenant','20000000-0000-4000-8000-000000000002','system_resource_layout',1);
INSERT INTO platform.cutover_groups VALUES
 ('tenant','10000000-0000-4000-8000-000000000001','iam_system_resources',1),
 ('tenant','20000000-0000-4000-8000-000000000002','iam_system_resources',1);
INSERT INTO platform.cutover_group_members VALUES
 ('tenant','10000000-0000-4000-8000-000000000001','iam_system_resources',1,'iam'),
 ('tenant','10000000-0000-4000-8000-000000000001','iam_system_resources',2,'system_resource_layout'),
 ('tenant','20000000-0000-4000-8000-000000000002','iam_system_resources',1,'iam'),
 ('tenant','20000000-0000-4000-8000-000000000002','iam_system_resources',2,'system_resource_layout');`

func prepareMaintenancePG18Schemas(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, maintenancePG18SchemaSQL); err != nil {
		t.Fatalf("prepare maintenance schemas: %v", err)
	}
	if origin == "v1_adopted" {
		seedMaintenancePG18V1Snapshot(t, ctx, database)
	}
}

func execMaintenancePG18Tenant(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant maintenancePG18Tenant, statement string, arguments ...any,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		tenant.schema+", pg_catalog"); err == nil {
		_, err = transaction.ExecContext(ctx, statement, arguments...)
	}
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("execute in %s: %v", tenant.schema, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
