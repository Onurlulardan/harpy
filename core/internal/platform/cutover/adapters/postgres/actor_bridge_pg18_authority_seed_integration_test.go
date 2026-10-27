//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func seedActorBridgeAuthority(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	for _, tenant := range actorBridgeTenants {
		_, err := admin.Exec(ctx, `
INSERT INTO platform.domain_cutovers (
    scope_kind, scope_id, domain, state, writer, generation,
    requested_by, reason, change_reference
) VALUES
    ('tenant',$1,'iam','v2','v2',7,$2,'actor bridge PG18','IAM-BRIDGE-1'),
    ('tenant',$1,'system_resource_layout','v2','v2',8,$2,
     'actor bridge PG18','IAM-BRIDGE-1');
INSERT INTO platform.cutover_groups (
    scope_kind, scope_id, group_id, state, writer, generation,
    requested_by, reason, change_reference
) VALUES (
    'tenant',$1,'iam_system_resources','v2','v2',9,$2,
    'actor bridge PG18','IAM-BRIDGE-1'
);
INSERT INTO platform.cutover_group_members (
    scope_kind, scope_id, group_id, ordinal, domain
) VALUES
    ('tenant',$1,'iam_system_resources',1,'iam'),
	    ('tenant',$1,'iam_system_resources',2,'system_resource_layout');`,
			pgx.QueryExecModeSimpleProtocol,
			tenant.id, integrationOwnerID)
		if err != nil {
			t.Fatalf("seed actor bridge authority for %s: %v", tenant.schema, err)
		}
	}
}

func seedActorBridgeLegacySentinels(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	for _, tenant := range actorBridgeTenants {
		statement := `CREATE TABLE ` + pgx.Identifier{tenant.schema, "users_form_table"}.Sanitize() + ` (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    email text NOT NULL
);
INSERT INTO ` + pgx.Identifier{tenant.schema, "users_form_table"}.Sanitize() + `
    (id, tenant_id, email)
VALUES ('88000000-0000-4000-8000-000000000001',$1,'legacy@example.com')`
		if _, err := admin.Exec(ctx, statement,
			pgx.QueryExecModeSimpleProtocol, tenant.id); err != nil {
			t.Fatalf("seed V1 sentinel in %s: %v", tenant.schema, err)
		}
	}
}

func assertActorBridgeLegacySentinels(
	t *testing.T, ctx context.Context, admin *pgx.Conn, origin string,
) {
	t.Helper()
	if origin != "v1_adopted" {
		return
	}
	for _, tenant := range actorBridgeTenants {
		query := `SELECT tenant_id::text || ':' || email FROM ` +
			pgx.Identifier{tenant.schema, "users_form_table"}.Sanitize() +
			` WHERE id='88000000-0000-4000-8000-000000000001'`
		var sentinel string
		if err := admin.QueryRow(ctx, query).Scan(&sentinel); err != nil ||
			sentinel != tenant.id+":legacy@example.com" {
			t.Fatalf("V1 sentinel in %s = %q, err=%v", tenant.schema, sentinel, err)
		}
	}
}
