//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	layoutScopeID = "30000000-0000-4000-8000-000000000001"
	layoutActorID = "30000000-0000-4000-8000-000000000002"
)

func TestSystemResourceLayoutCutoverDomainAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	database := startCutoverIntegrationDatabase(t, ctx)
	seedSystemResourceLayoutAuthority(t, ctx, database)
	assertSystemResourceLayoutGroup(t, ctx, database)
	assertCutoverGenerationAuthorityKeys(t, ctx, database)
	assertSystemResourceLayoutRetirementRoutine(t, ctx, database)
	assertSystemResourceLayoutDomainNegatives(t, ctx, database)
}

func assertCutoverGenerationAuthorityKeys(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `
CREATE TABLE platform.integration_generation_authority_probe (
    scope_kind text NOT NULL, scope_id uuid NOT NULL,
    group_id text NOT NULL, group_generation bigint NOT NULL,
    domain text NOT NULL, domain_generation bigint NOT NULL,
    CONSTRAINT generation_probe_group_fk FOREIGN KEY (
        scope_kind,scope_id,group_id,group_generation
    ) REFERENCES platform.cutover_groups (
        scope_kind,scope_id,group_id,generation
    ),
    CONSTRAINT generation_probe_domain_fk FOREIGN KEY (
        scope_kind,scope_id,domain,domain_generation
    ) REFERENCES platform.domain_cutovers (
        scope_kind,scope_id,domain,generation
    )
);
INSERT INTO platform.integration_generation_authority_probe VALUES (
    'tenant','30000000-0000-4000-8000-000000000001',
    'iam_system_resources',1,'system_resource_layout',1
);`, pgx.QueryExecModeSimpleProtocol)
	if err != nil {
		t.Fatalf("create and seed generation authority probe: %v", err)
	}
	_, err = database.admin.Exec(ctx, `INSERT INTO platform.integration_generation_authority_probe VALUES (
    'tenant','30000000-0000-4000-8000-000000000001',
    'iam_system_resources',2,'system_resource_layout',1)`)
	assertForeignKeyViolation(t, err, "generation_probe_group_fk")
	_, err = database.admin.Exec(ctx, `INSERT INTO platform.integration_generation_authority_probe VALUES (
    'tenant','30000000-0000-4000-8000-000000000001',
    'iam_system_resources',1,'system_resource_layout',2)`)
	assertForeignKeyViolation(t, err, "generation_probe_domain_fk")
}

func seedSystemResourceLayoutAuthority(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `
INSERT INTO platform.domain_cutovers (
    scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference
) VALUES
    ('tenant',$1,'iam','shadow','v1',1,$2,'layout fixture','IAM-LAYOUT-1'),
    ('tenant',$1,'system_resource_layout','shadow','v1',1,$2,'layout fixture','IAM-LAYOUT-1');
INSERT INTO platform.cutover_groups (
    scope_kind,scope_id,group_id,state,writer,generation,requested_by,reason,change_reference
) VALUES (
    'tenant',$1,'iam_system_resources','shadow','v1',1,$2,'layout fixture','IAM-LAYOUT-1'
);
INSERT INTO platform.cutover_group_members (
    scope_kind,scope_id,group_id,ordinal,domain
) VALUES
    ('tenant',$1,'iam_system_resources',1,'iam'),
    ('tenant',$1,'iam_system_resources',2,'system_resource_layout');
INSERT INTO platform.cutover_transition_events (
    event_id,target_kind,event_kind,scope_kind,scope_id,group_id,domain,
    previous_state,new_state,previous_writer,new_writer,previous_generation,
    new_generation,actor_id,requested_by,reason,change_reference
) VALUES (
    '30000000-0000-4000-8000-000000000003','domain','request','tenant',$1,
    'iam_system_resources','system_resource_layout','shadow','quiescing','v1','none',
    1,1,$2,$2,'layout transition request','IAM-LAYOUT-2'
);`, pgx.QueryExecModeSimpleProtocol, layoutScopeID, layoutActorID)
	if err != nil {
		t.Fatalf("seed exact IAM system-resource authority: %v", err)
	}
}

func assertSystemResourceLayoutGroup(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	var members []string
	err := database.admin.QueryRow(ctx, `
SELECT pg_catalog.array_agg(member.domain ORDER BY member.ordinal)
FROM platform.cutover_group_members AS member
JOIN platform.domain_cutovers AS domain_row
  ON domain_row.scope_kind=member.scope_kind AND domain_row.scope_id=member.scope_id
 AND domain_row.domain=member.domain
WHERE member.scope_kind='tenant' AND member.scope_id=$1
  AND member.group_id='iam_system_resources'`, layoutScopeID).Scan(&members)
	if err != nil || len(members) != 2 || members[0] != "iam" ||
		members[1] != "system_resource_layout" {
		t.Fatalf("exact IAM system-resource members = %v, err=%v", members, err)
	}
}

func assertSystemResourceLayoutRetirementRoutine(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	migrator := openCutoverMigrator(t, ctx, database.dsn)
	defer migrator.Close(ctx)
	if _, err := migrator.Exec(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatalf("set owner role for retirement provision: %v", err)
	}
	var revision int64
	err := migrator.QueryRow(ctx,
		`SELECT platform.contract_retirement_provision_v1($1,$2,$3,0)`,
		"rq.alias."+strings.Repeat("a", 64), "system_resource_layout",
		"system_resource_layout").Scan(&revision)
	if err != nil || revision != 1 {
		t.Fatalf("layout retirement provision revision = %d, err=%v", revision, err)
	}
}

func assertSystemResourceLayoutDomainNegatives(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	statement := `INSERT INTO platform.domain_cutovers (
scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference
) VALUES ($1,$2,$3,'v1','v1',1,$4,'negative','IAM-LAYOUT-NEGATIVE')`
	_, err := database.admin.Exec(ctx, statement, "tenant", layoutScopeID,
		"unknown_layout", layoutActorID)
	assertConstraintViolation(t, err, "domain_cutovers_domain_check")
	_, err = database.admin.Exec(ctx, statement, "global", layoutScopeID,
		"system_resource_layout", layoutActorID)
	assertConstraintViolation(t, err, "domain_cutovers_global_check")
}

func assertConstraintViolation(t *testing.T, err error, constraint string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23514" ||
		postgresError.ConstraintName != constraint {
		t.Fatalf("error = %v, want 23514 on %s", err, constraint)
	}
}

func assertForeignKeyViolation(t *testing.T, err error, constraint string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23503" ||
		postgresError.ConstraintName != constraint {
		t.Fatalf("error = %v, want 23503 on %s", err, constraint)
	}
}
