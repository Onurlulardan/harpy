//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const iamPlatformGuardSchema = `
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE TABLE platform.tenants (
    id uuid PRIMARY KEY, schema_name name NOT NULL UNIQUE
);
CREATE TABLE platform.environment_identity (
    singleton boolean PRIMARY KEY CHECK (singleton),
    environment_id uuid NOT NULL UNIQUE
);
CREATE TABLE platform.domain_cutovers (
    scope_kind text NOT NULL, scope_id uuid NOT NULL, domain text NOT NULL,
    state text NOT NULL, writer text NOT NULL, generation bigint NOT NULL,
    PRIMARY KEY (scope_kind,scope_id,domain)
);
CREATE TABLE platform.cutover_groups (
    scope_kind text NOT NULL, scope_id uuid NOT NULL, group_id text NOT NULL,
    state text NOT NULL, writer text NOT NULL, generation bigint NOT NULL,
    PRIMARY KEY (scope_kind,scope_id,group_id)
);
CREATE TABLE platform.cutover_group_members (
    scope_kind text NOT NULL, scope_id uuid NOT NULL, group_id text NOT NULL,
    ordinal smallint NOT NULL, domain text NOT NULL,
    PRIMARY KEY (scope_kind,scope_id,group_id,ordinal),
    FOREIGN KEY (scope_kind,scope_id,group_id)
        REFERENCES platform.cutover_groups (scope_kind,scope_id,group_id),
    FOREIGN KEY (scope_kind,scope_id,domain)
        REFERENCES platform.domain_cutovers (scope_kind,scope_id,domain),
    UNIQUE (scope_kind,scope_id,group_id,domain)
);
INSERT INTO platform.tenants VALUES
    ('10000000-0000-4000-8000-000000000001','t_iam_a'),
    ('20000000-0000-4000-8000-000000000002','t_iam_b');
INSERT INTO platform.environment_identity VALUES
    (true,'30000000-0000-4000-8000-000000000001');
INSERT INTO platform.cutover_groups VALUES
    ('tenant','10000000-0000-4000-8000-000000000001',
     'iam_system_resources','v2','v2',1),
    ('tenant','20000000-0000-4000-8000-000000000002',
     'iam_system_resources','v2','v2',1);
INSERT INTO platform.domain_cutovers VALUES
    ('tenant','10000000-0000-4000-8000-000000000001','iam','v2','v2',1),
    ('tenant','10000000-0000-4000-8000-000000000001',
     'system_resource_layout','v2','v2',1),
    ('tenant','20000000-0000-4000-8000-000000000002','iam','v2','v2',1),
    ('tenant','20000000-0000-4000-8000-000000000002',
     'system_resource_layout','v2','v2',1);
INSERT INTO platform.cutover_group_members VALUES
    ('tenant','10000000-0000-4000-8000-000000000001',
     'iam_system_resources',1,'iam'),
    ('tenant','10000000-0000-4000-8000-000000000001',
     'iam_system_resources',2,'system_resource_layout'),
    ('tenant','20000000-0000-4000-8000-000000000002',
     'iam_system_resources',1,'iam'),
    ('tenant','20000000-0000-4000-8000-000000000002',
     'iam_system_resources',2,'system_resource_layout');`

func prepareIAMPlatformGuardSchema(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	origin string,
) {
	t.Helper()
	connection, err := database.migrator.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, iamPlatformGuardSchema); err != nil {
		t.Fatalf("prepare IAM platform guard schema: %v", err)
	}
	base := []string{
		"20260716010000_create_iam_login_directory",
		"20260716012000_create_iam_refresh_locators",
		"20260716012010_create_iam_auth_throttle_buckets",
	}
	for _, name := range append(base, platformAuthGuardMigrations...) {
		applyIAMPlatformGuardDocument(t, ctx, connection,
			name, readIAMPlatformAuthMigration(t, origin, name))
	}
}

func applyIAMPlatformGuardDocument(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	name string, content []byte,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin %s: %v", name, err)
	}
	if _, err := transaction.ExecContext(ctx, string(content)); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("apply %s: %v", name, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit %s: %v", name, err)
	}
}
