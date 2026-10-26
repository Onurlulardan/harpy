//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const iamPlatformAuthPrerequisites = `
CREATE TABLE platform.environment_identity (
    environment_id uuid PRIMARY KEY
);
CREATE TABLE platform.domain_cutovers (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    domain text NOT NULL,
    generation bigint NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, domain),
    UNIQUE (scope_kind, scope_id, domain, generation)
);
CREATE TABLE platform.cutover_groups (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    generation bigint NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id),
    UNIQUE (scope_kind, scope_id, group_id, generation)
);
CREATE TABLE platform.cutover_group_members (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    ordinal smallint NOT NULL,
    domain text NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id, ordinal),
    FOREIGN KEY (scope_kind, scope_id, group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id),
    FOREIGN KEY (scope_kind, scope_id, domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain),
    UNIQUE (scope_kind, scope_id, group_id, domain)
);
INSERT INTO platform.environment_identity (environment_id)
VALUES ('30000000-0000-4000-8000-000000000001');
INSERT INTO platform.domain_cutovers (scope_kind, scope_id, domain, generation)
VALUES
    ('tenant', '10000000-0000-4000-8000-000000000001', 'iam', 1),
    ('tenant', '10000000-0000-4000-8000-000000000001',
     'system_resource_layout', 1),
    ('tenant', '20000000-0000-4000-8000-000000000002', 'iam', 1),
    ('tenant', '20000000-0000-4000-8000-000000000002',
     'system_resource_layout', 1);
INSERT INTO platform.cutover_groups (scope_kind, scope_id, group_id, generation)
VALUES
    ('tenant', '10000000-0000-4000-8000-000000000001',
     'iam_system_resources', 1),
    ('tenant', '20000000-0000-4000-8000-000000000002',
     'iam_system_resources', 1);
INSERT INTO platform.cutover_group_members (
    scope_kind, scope_id, group_id, ordinal, domain
) VALUES
    ('tenant', '10000000-0000-4000-8000-000000000001',
     'iam_system_resources', 1, 'iam'),
    ('tenant', '10000000-0000-4000-8000-000000000001',
     'iam_system_resources', 2, 'system_resource_layout'),
    ('tenant', '20000000-0000-4000-8000-000000000002',
     'iam_system_resources', 1, 'iam'),
    ('tenant', '20000000-0000-4000-8000-000000000002',
     'iam_system_resources', 2, 'system_resource_layout');`

func prepareIAMPlatformAuthPrerequisites(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, iamPlatformAuthPrerequisites); err != nil {
		t.Fatalf("prepare IAM platform auth prerequisites: %v", err)
	}
}
