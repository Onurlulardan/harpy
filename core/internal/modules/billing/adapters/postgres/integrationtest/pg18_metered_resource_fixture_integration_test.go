//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type meteredResourceFixture struct {
	database    *sql.DB
	authorityID uuid.UUID
	tenantID    uuid.UUID
	actorID     uuid.UUID
	periodID    uuid.UUID
	now         time.Time
}

func prepareMeteredResourceFixture(
	t *testing.T,
	ctx context.Context,
	origin string,
) meteredResourceFixture {
	t.Helper()
	database := prepareBillingOrigin(t, ctx, origin)
	globalAuthority := billingGlobalAuthorityID(t, ctx, database)
	connection := billingRuntimeConnection(t, ctx, database)
	tenant := prepareBillingTenantFixture(t, ctx, connection, globalAuthority)
	fixture := meteredResourceFixture{
		database: database, authorityID: tenant.authorityID,
		tenantID: uuid.MustParse(billingTenantID), actorID: billingV7(t),
		periodID: billingV7(t), now: tenant.now,
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, grant := range []struct {
		key       string
		valueType string
		boolean   any
		integer   any
	}{
		{"api.builder.enabled", "boolean", true, nil},
		{"api.active_endpoints", "integer", nil, int64(1)},
		{"analytics.dashboard_builder.enabled", "boolean", true, nil},
		{"analytics.active_dashboards", "integer", nil, int64(1)},
	} {
		_, err = transaction.ExecContext(ctx, `
SELECT platform.billing_add_offering_grant_v1(
 $1,$2,$3,$4,$5,$6,false,$7
)`, globalAuthority, tenant.offeringVersionID, grant.key, grant.valueType,
			grant.boolean, grant.integer, fixture.now)
		if err != nil {
			break
		}
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_meter_periods(
 id,authority_id,tenant_id,entitlement_definition_id,entitlement_key,meter_kind,
 reset_policy,period_start,period_end,state,closed_at,revision
) VALUES($1,$2,$3,$4,'api.active_endpoints','gauge','never',$5,$6,'open',NULL,1)
`, fixture.periodID, fixture.authorityID, fixture.tenantID,
			uuid.MustParse("01900000-0000-7000-8000-00000000000d"),
			fixture.now.Add(-time.Hour), fixture.now.AddDate(100, 0, 0))
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_usage_totals(
 authority_id,tenant_id,meter_period_id,entitlement_key,meter_kind,
 committed_total,reserved_total,gauge_total,revision,updated_at
) VALUES($1,$2,$3,'api.active_endpoints','gauge',0,0,0,1,$4)
`, fixture.authorityID, fixture.tenantID, fixture.periodID,
			fixture.now.Add(-time.Hour))
	}
	var assignmentID uuid.UUID
	if err == nil {
		err = transaction.QueryRowContext(ctx, `
SELECT id FROM platform.billing_tenant_base_offerings
WHERE authority_id=$1 AND tenant_id=$2 AND offering_version_id=$3
`, fixture.authorityID, fixture.tenantID, tenant.offeringVersionID).Scan(&assignmentID)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
UPDATE platform.tenancy_deployment_policy
SET deployment_mode='saas',maximum_customer_tenants=100
`)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
UPDATE platform.tenancy_tenant_bindings SET purpose='saas_customer'
WHERE tenant_id=$1
`, fixture.tenantID)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_installation_base_assignments(
 operation_id,environment_id,tenant_id,authority_id,assignment_id,deployment_mode,
 offering_version_id,offering_revision,offering_sha256,offering_effective_at,
 source_sequence,source_sha256,manifest_sha256,assigned_at
) VALUES($2,$3,$1,$4,$5,'saas',$6,1,$7,$8,NULL,NULL,$9,$8)
`, fixture.tenantID, billingV7(t), uuid.MustParse(billingEnvironmentID),
			fixture.authorityID, assignmentID, tenant.offeringVersionID,
			billingDigestHex("metered-offering"), fixture.now.Add(-time.Hour),
			billingDigestHex("metered-manifest"))
	}
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("prepare metered resource fixture: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func billingDigestHex(value string) string {
	return fmt.Sprintf("%x", billingDigest(value))
}
