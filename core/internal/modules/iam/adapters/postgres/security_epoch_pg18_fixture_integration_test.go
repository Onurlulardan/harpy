//go:build integration

package postgres

const securityEpochPG18Fixture = `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime NOLOGIN NOINHERIT;
GRANT harpy_runtime TO harpy_v2_runtime;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA platform TO harpy_runtime;
CREATE TABLE platform.tenancy_deployment_policy (
    environment_id uuid PRIMARY KEY
);
CREATE TABLE platform.tenancy_tenant_bindings (
    tenant_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL REFERENCES platform.tenancy_deployment_policy
);
INSERT INTO platform.tenancy_deployment_policy VALUES
    ('30000000-0000-4000-8000-000000000003');
INSERT INTO platform.tenancy_tenant_bindings VALUES
    ('10000000-0000-4000-8000-000000000001',
     '30000000-0000-4000-8000-000000000003');
CREATE SCHEMA t_epoch AUTHORIZATION harpy_owner;
CREATE TABLE t_epoch.tenancy_schema_identity (
    singleton boolean PRIMARY KEY CHECK (singleton), tenant_id uuid NOT NULL
);
INSERT INTO t_epoch.tenancy_schema_identity VALUES
    (true, '10000000-0000-4000-8000-000000000001');
CREATE TABLE t_epoch.iam_sessions (
    id uuid PRIMARY KEY, user_id uuid NOT NULL, state text NOT NULL,
    revision bigint NOT NULL, updated_at timestamptz NOT NULL,
    revoked_at timestamptz, retain_until timestamptz NOT NULL
);
CREATE TABLE t_epoch.iam_refresh_families (
    id uuid PRIMARY KEY, session_id uuid NOT NULL, state text NOT NULL,
    revoke_reason text, revision bigint NOT NULL,
    updated_at timestamptz NOT NULL, revoked_at timestamptz
);
CREATE TABLE t_epoch.iam_refresh_tokens (
    id uuid PRIMARY KEY, family_id uuid NOT NULL, state text NOT NULL,
    revision bigint NOT NULL, updated_at timestamptz NOT NULL,
    revoked_at timestamptz
);
CREATE TABLE t_epoch.iam_one_time_tokens (
    id uuid PRIMARY KEY, state text NOT NULL, revision bigint NOT NULL,
    updated_at timestamptz NOT NULL, revoked_at timestamptz,
    purge_after timestamptz NOT NULL
);
CREATE TABLE t_epoch.iam_auth_challenges (
    id uuid PRIMARY KEY, state text NOT NULL, revision bigint NOT NULL,
    updated_at timestamptz NOT NULL, terminal_at timestamptz,
    purge_after timestamptz NOT NULL
);`
