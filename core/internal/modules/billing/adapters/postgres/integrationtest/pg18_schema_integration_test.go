//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

const billingEnvironmentID = "01900000-0000-7000-8000-000000000007"
const billingTenantID = "71000000-0000-4000-8000-000000000071"

func prepareBillingPrerequisites(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	const statement = `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT PASSWORD 'runtime-integration-only';
CREATE ROLE harpy_adopter LOGIN NOINHERIT PASSWORD 'adopter-integration-only';
CREATE ROLE harpy_cutover_decoder NOLOGIN NOINHERIT NOSUPERUSER
NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0;
CREATE ROLE harpy_cutover_authority_owner NOLOGIN;
CREATE ROLE harpy_cutover_finalizer LOGIN NOINHERIT;
GRANT harpy_runtime TO harpy_v2_runtime WITH INHERIT FALSE, SET TRUE;
GRANT harpy_owner TO harpy_v2_runtime;
GRANT harpy_owner TO harpy_adopter;
GRANT harpy_cutover_authority_owner TO harpy_cutover_finalizer
    WITH INHERIT FALSE, SET TRUE;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA platform TO harpy_cutover_authority_owner;
CREATE TABLE platform.environment_identity (
    environment_id uuid PRIMARY KEY
);
CREATE TABLE platform.tenancy_deployment_policy (
    singleton boolean PRIMARY KEY CHECK (singleton),
    environment_id uuid NOT NULL UNIQUE,
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('onpremise','saas')),
    maximum_customer_tenants bigint,
    created_at timestamptz NOT NULL
);
CREATE TABLE platform.tenants (
    id uuid PRIMARY KEY,
    schema_name name NOT NULL UNIQUE,
    status text NOT NULL,
    deleted_at timestamptz
);
CREATE TABLE platform.tenancy_tenant_bindings (
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    purpose text NOT NULL,
    PRIMARY KEY (environment_id, tenant_id)
);
-- Billing owns only the FK source. This minimal dependency proves its owner
-- stream composes after the already-certified platform_async jobs migration.
CREATE TABLE platform.jobs (
    id uuid PRIMARY KEY
);
INSERT INTO platform.environment_identity VALUES ('` + billingEnvironmentID + `');
INSERT INTO platform.tenancy_deployment_policy VALUES (
    true, '` + billingEnvironmentID + `', 'onpremise', 0, statement_timestamp()
);
INSERT INTO platform.tenants VALUES (
    '` + billingTenantID + `', 't_billing', 'ACTIVE', NULL
);
INSERT INTO platform.tenancy_tenant_bindings VALUES (
    '` + billingEnvironmentID + `', '` + billingTenantID + `', 'onpremise_primary'
);`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare Billing prerequisites: %v", err)
	}
	applyBillingCutoverPlatformMigrations(t, ctx, database, origin)
	global := newBillingCutoverAuthorityFixture(
		"global", uuid.MustParse(billingEnvironmentID), "billing",
		1, "v2", "v2", billingAcceptanceA, "billing",
	)
	if err := seedBillingCutoverAuthority(ctx, database, global); err != nil {
		t.Fatalf("seed global Billing cutover prerequisite: %v", err)
	}
	tenant := newBillingCutoverAuthorityFixture(
		"tenant", uuid.MustParse(billingTenantID), "billing_forms",
		1, "v2", "v2", billingAcceptanceA, "billing", "forms",
	)
	if err := seedBillingCutoverAuthority(ctx, database, tenant); err != nil {
		t.Fatalf("seed Billing cutover prerequisites: %v", err)
	}
	prepareBillingAuthorityCapabilityFixture(t, ctx, database)
	prepareBillingContractRetirementFixture(t, ctx, database)
}
