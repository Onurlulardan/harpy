//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareCutoverOwnerPrerequisiteSurfaces(
	t *testing.T, ctx context.Context, connection *pgx.Conn,
) {
	t.Helper()
	prepareRuntimeSnapshotAuthoritySurfaces(t, ctx, connection)
	prepareGen70ActivationFenceSurfaces(t, ctx, connection)
}

func prepareRuntimeSnapshotAuthoritySurfaces(
	t *testing.T, ctx context.Context, connection *pgx.Conn,
) {
	t.Helper()
	const statement = `
CREATE TABLE platform.billing_global_authorities (id bigint);
CREATE TABLE platform.billing_tenant_authorities (id bigint);
CREATE TABLE platform.billing_offerings (id bigint);
CREATE TABLE platform.billing_offering_versions (id bigint);
CREATE TABLE platform.billing_price_versions (id bigint);
CREATE TABLE platform.billing_subscriptions (id bigint);
CREATE TABLE platform.billing_usage_totals (id bigint);
CREATE TABLE platform.billing_discounts (id bigint);
REVOKE ALL ON TABLE
    platform.billing_global_authorities,
    platform.billing_tenant_authorities,
    platform.billing_offerings,
    platform.billing_offering_versions,
    platform.billing_price_versions,
    platform.billing_subscriptions,
    platform.billing_usage_totals,
    platform.billing_discounts
FROM PUBLIC, harpy_runtime, harpy_v2_runtime;

CREATE VIEW platform.billing_catalog_offering_snapshot_read_v1
WITH (security_barrier=true) AS SELECT 1::bigint AS id;
CREATE VIEW platform.billing_catalog_campaign_snapshot_read_v1
WITH (security_barrier=true) AS SELECT 1::bigint AS id;
REVOKE ALL ON TABLE
    platform.billing_catalog_offering_snapshot_read_v1,
    platform.billing_catalog_campaign_snapshot_read_v1
FROM PUBLIC, harpy_v2_runtime;
GRANT SELECT ON
    platform.billing_catalog_offering_snapshot_read_v1,
    platform.billing_catalog_campaign_snapshot_read_v1
TO harpy_runtime;

DO $fixture$
DECLARE
    v_signature text;
BEGIN
    FOREACH v_signature IN ARRAY ARRAY[
        'platform.billing_resolve_current_tenant_authority_v1(uuid)',
        'platform.billing_resolve_current_global_authority_v1(text)',
        'platform.billing_effective_subscription_price_snapshot_read_v1(uuid,uuid,uuid,timestamptz)',
        'platform.billing_subscription_snapshot_read_v1(uuid,uuid)',
        'platform.billing_usage_context_snapshot_read_v1(uuid,uuid)',
        'platform.billing_usage_snapshot_read_v1(uuid,uuid,uuid,uuid,uuid)',
        'platform.billing_discount_eligibility_snapshot_read_v1(uuid,uuid,text,uuid,timestamptz,uuid)',
        'platform.billing_discount_offerings_snapshot_read_v1(uuid,uuid)'
    ] LOOP
        EXECUTE pg_catalog.format(
            'CREATE FUNCTION %s RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS %L',
            v_signature, 'BEGIN RETURN; END'
        );
        EXECUTE pg_catalog.format(
            'REVOKE ALL ON FUNCTION %s FROM PUBLIC, harpy_v2_runtime',
            v_signature
        );
        EXECUTE pg_catalog.format(
            'GRANT EXECUTE ON FUNCTION %s TO harpy_runtime', v_signature
        );
    END LOOP;
    FOREACH v_signature IN ARRAY ARRAY[
        'platform.workflow_resolve_current_tenant_authority_v1(uuid)',
        'platform.analytics_resolve_current_tenant_authority_v1(uuid)'
    ] LOOP
        EXECUTE pg_catalog.format(
            'CREATE FUNCTION %s RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS %L',
            v_signature, 'BEGIN RETURN; END'
        );
        EXECUTE pg_catalog.format(
            'REVOKE ALL ON FUNCTION %s FROM PUBLIC, harpy_runtime, harpy_v2_runtime',
            v_signature
        );
    END LOOP;
END
$fixture$;`
	if _, err := connection.Exec(ctx, statement); err != nil {
		t.Fatalf("prepare runtime snapshot authority surfaces: %v", err)
	}
}
