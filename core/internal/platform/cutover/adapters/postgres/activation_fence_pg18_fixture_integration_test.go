//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareGen70ActivationFenceSurfaces(
	t *testing.T, ctx context.Context, connection *pgx.Conn,
) {
	t.Helper()
	const statement = `
CREATE TABLE platform.tenancy_activation_step_capabilities (id bigint);
CREATE TABLE platform.platform_async_installation_activation_receipts (id bigint);
CREATE TABLE platform.tenancy_activation_command_capabilities (id bigint);
CREATE TABLE platform.tenancy_authority_reconcile_scopes (id bigint);
CREATE TABLE platform.billing_installation_assignment_contexts (id bigint);
REVOKE ALL ON TABLE
    platform.tenancy_activation_step_capabilities,
    platform.platform_async_installation_activation_receipts,
    platform.tenancy_activation_command_capabilities,
    platform.tenancy_authority_reconcile_scopes,
    platform.billing_installation_assignment_contexts
FROM PUBLIC, harpy_runtime, harpy_v2_runtime;
DO $fixture$
DECLARE
    v_signature text;
BEGIN
    FOREACH v_signature IN ARRAY ARRAY[
        'platform.tenancy_onboarding_activate_v3(uuid,text,bigint)',
        'platform.tenancy_onboarding_activate_v2(uuid,text,bigint)',
        'platform.tenancy_activation_initialize_v5(uuid,uuid,uuid,text,text,uuid,text,text,text,text)',
        'platform.tenancy_authority_reconcile_current_v1(uuid,text)',
        'platform.billing_installation_assign_v3(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)',
        'platform.iam_installation_activation_assert_v2(uuid,uuid,uuid,name)',
        'platform.iam_installation_security_epoch_initialize_v2(uuid,uuid,uuid,name,timestamptz)',
        'platform.iam_installation_identity_directory_upsert_v1(uuid,uuid,uuid,name,bigint,bigint,uuid,uuid,text,bytea,bigint,bytea)',
        'platform.iam_installation_instance_owner_bind_v2(uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text)',
        'platform.billing_installation_open_usage_v2(uuid,uuid,text)',
        'platform.platformops_installation_recovery_initialize_v2(uuid,uuid,uuid,name,text,bigint,text,text)',
        'platform.platform_async_initial_admin_setup_enqueue_v2(uuid,uuid,uuid,uuid,text,name,text)',
        'platform.platform_async_installation_activation_event_v2(uuid,uuid,uuid,uuid,name,text)',
        'platform.platformauthority_installation_owner_bind_v3(uuid,uuid,uuid,name,text,uuid,uuid,uuid,text)',
        'platform.tenancy_onboarding_complete_v2(uuid,uuid,text,text,bigint)',
        'platform.platformauthority_installation_head_pin_v3(uuid,uuid,uuid,name,uuid,text,uuid,bigint,text,text,text,text,text)',
        'platform.tenancy_onboarding_accept_v2(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)'
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
        'platform.tenancy_activation_step_assert_v1(uuid,text)',
        'platform.tenancy_activation_step_consume_v1(uuid,text)',
        'platform.tenancy_activation_step_context_assert_v1(uuid,text,uuid,uuid,name,text)',
        'platform.tenancy_activation_step_commit_guard_v1()',
        'platform.platform_async_installation_activation_evidence_assert_v1(uuid,uuid)',
        'platform.tenancy_activation_command_consume_v1(uuid,text,uuid,uuid,text,text,text,text,uuid,bigint,text,timestamptz,bigint,text)',
        'platform.tenancy_authority_reconcile_scope_assert_v1(text)',
        'platform.tenancy_activation_capability_commit_guard_v1()',
        'platform.tenancy_authority_scope_commit_guard_v1()',
        'platform.billing_installation_id_v1(uuid,text)',
        'platform.analytics_reconcile_current_tenant_authorities_v2()',
        'platform.workflow_reconcile_current_tenant_authorities_v2()',
        'platform.billing_reconcile_current_global_authority_v2()',
        'platform.billing_reconcile_current_tenant_authorities_v2()'
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
    FOREACH v_signature IN ARRAY ARRAY[
        'platform.tenancy_onboarding_activate_v1(uuid,text,bigint)',
        'platform.tenancy_activation_initialize_v1(uuid,uuid,uuid,text,text,uuid,text,text,text)',
        'platform.tenancy_activation_initialize_v2(uuid,uuid,uuid,text,text,uuid,text,text,text)',
        'platform.tenancy_activation_initialize_v3(uuid,uuid,uuid,text,text,uuid,text,text,text)',
        'platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text)',
        'platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)',
        'platform.billing_reconcile_current_global_authority_v1()',
        'platform.billing_reconcile_current_tenant_authorities_v1()',
        'platform.billing_register_global_authority_v1(uuid,bigint,bigint,timestamptz)',
        'platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,bigint,timestamptz)',
        'platform.platformauthority_installation_head_pin_v2(uuid,uuid,uuid,name,uuid,text,uuid,bigint,text,text,text,text,text)',
        'platform.tenancy_activation_initialize_v4(uuid,uuid,uuid,text,text,uuid,text,text,text,text)',
        'platform.tenancy_onboarding_complete_v1(uuid,uuid,text,text,bigint)',
        'platform.tenancy_onboarding_accept_v1(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)',
        'platform.billing_installation_open_usage_v1(uuid,uuid,text)',
        'platform.platformops_installation_recovery_initialize_v1(uuid,uuid,uuid,name,text,bigint,text,text)',
        'platform.platform_async_initial_admin_setup_enqueue_v1(uuid,uuid,uuid,text,name,text)',
        'platform.platform_async_installation_activation_event_v1(uuid,uuid,uuid,uuid,name,text)',
        'platform.platformauthority_installation_owner_bind_v1(uuid,uuid,uuid,name,text,uuid,uuid,uuid,text)',
        'platform.platformauthority_installation_owner_bind_v2(uuid,uuid,uuid,name,text,uuid,uuid,uuid,text)',
        'platform.platformauthority_installation_head_pin_v1(uuid,uuid,uuid,name,uuid,text,uuid,bigint,text,text,text,text,text)',
        'platform.tenancy_activation_assert_v1(uuid,uuid,name)',
        'platform.tenancy_activation_assert_v2(uuid,uuid,name)',
        'platform.iam_security_epoch_initialize_v1(uuid,timestamptz)',
        'platform.iam_installation_instance_owner_bind_v1(uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text)',
        'platform.iam_installation_instance_owner_input_v1(uuid)',
        'platform.analytics_reconcile_current_tenant_authorities_v1()',
        'platform.analytics_register_tenant_authority_v1(uuid,uuid,bigint,bigint,timestamptz)',
        'platform.workflow_reconcile_current_tenant_authorities_v1()',
        'platform.workflow_register_tenant_authority_v1(uuid,uuid,bigint,bigint,timestamptz)'
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
END
$fixture$;`
	if _, err := connection.Exec(ctx, statement); err != nil {
		t.Fatalf("prepare Gen70 activation fence surfaces: %v", err)
	}
}
