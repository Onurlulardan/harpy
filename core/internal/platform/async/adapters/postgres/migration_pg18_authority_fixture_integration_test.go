//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	asyncTenantA     = "11111111-1111-4111-8111-111111111111"
	asyncTenantB     = "22222222-2222-4222-8222-222222222222"
	asyncGlobalScope = "33333333-3333-4333-8333-333333333333"
	asyncActor       = "44444444-4444-4444-8444-444444444444"
	asyncCausation   = "55555555-5555-4555-8555-555555555555"
)

func prepareAsyncAuthorityFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, statement := range []string{asyncAuthorityTables, asyncAuthorityRows} {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("prepare async authority fixture: %v", err)
		}
	}
}

const asyncAuthorityTables = `
CREATE SCHEMA platform;
CREATE TABLE platform.tenants (id uuid PRIMARY KEY);
CREATE TABLE platform.domain_cutovers (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    domain text NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, domain)
);
CREATE TABLE platform.cutover_groups (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id)
);
CREATE TABLE platform.cutover_group_members (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    domain text NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id, domain)
);`

const asyncAuthorityRows = `
INSERT INTO platform.tenants (id) VALUES
    ('11111111-1111-4111-8111-111111111111'),
    ('22222222-2222-4222-8222-222222222222');
INSERT INTO platform.cutover_groups (scope_kind, scope_id, group_id) VALUES
    ('tenant', '11111111-1111-4111-8111-111111111111', 'operations'),
    ('tenant', '11111111-1111-4111-8111-111111111111', 'communications'),
    ('tenant', '22222222-2222-4222-8222-222222222222', 'operations'),
    ('global', '33333333-3333-4333-8333-333333333333', 'operations'),
    ('global', '33333333-3333-4333-8333-333333333333', 'communications');
INSERT INTO platform.domain_cutovers (scope_kind, scope_id, domain) VALUES
    ('tenant', '11111111-1111-4111-8111-111111111111', 'workflow'),
    ('tenant', '11111111-1111-4111-8111-111111111111', 'notifications'),
    ('tenant', '22222222-2222-4222-8222-222222222222', 'workflow'),
    ('global', '33333333-3333-4333-8333-333333333333', 'workflow'),
    ('global', '33333333-3333-4333-8333-333333333333', 'notifications');
INSERT INTO platform.cutover_group_members
    (scope_kind, scope_id, group_id, domain) VALUES
    ('tenant', '11111111-1111-4111-8111-111111111111', 'operations', 'workflow'),
    ('tenant', '11111111-1111-4111-8111-111111111111', 'communications', 'notifications'),
    ('tenant', '22222222-2222-4222-8222-222222222222', 'operations', 'workflow'),
    ('global', '33333333-3333-4333-8333-333333333333', 'operations', 'workflow'),
    ('global', '33333333-3333-4333-8333-333333333333', 'communications', 'notifications');`
