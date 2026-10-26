//go:build integration

package postgres

const analyticsDashboardExportMaintenanceRecoveryWrappersPG18 = `
CREATE FUNCTION tenant_a.test_analytics_maintenance_unknown_claim(
    p_kind text,p_schema name,p_export uuid,p_group text,p_group_generation bigint,
    p_cutover_generation bigint,p_job uuid,p_attempt uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_owner uuid,p_effect uuid,p_request bytea,
    p_unknown timestamptz,p_cleanup_owner uuid,p_observed timestamptz
) RETURNS TABLE(claim_state text,cleanup_generation bigint,lease_until timestamptz)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $f$
SELECT * FROM platform.platform_async_analytics_dashboard_export_maintenance_unknown_claim_v1(
    p_kind,'11111111-1111-4111-8111-111111111111',p_schema,p_export,p_group,
    p_group_generation,p_cutover_generation,p_job,p_attempt,p_attempt_number,
    p_lease_generation,p_original_owner,p_effect,p_request,p_unknown,
    p_cleanup_owner,p_observed) $f$;
CREATE FUNCTION tenant_a.test_analytics_maintenance_unknown_checkpoint(
    p_kind text,p_schema name,p_export uuid,p_group text,p_group_generation bigint,
    p_cutover_generation bigint,p_job uuid,p_attempt uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_owner uuid,p_effect uuid,p_request bytea,
    p_unknown timestamptz,p_cleanup_generation bigint,p_cleanup_owner uuid,
    p_observed timestamptz
) RETURNS timestamptz LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $f$
SELECT platform.platform_async_analytics_dashboard_export_maintenance_unknown_checkpoint_v1(
    p_kind,'11111111-1111-4111-8111-111111111111',p_schema,p_export,p_group,
    p_group_generation,p_cutover_generation,p_job,p_attempt,p_attempt_number,
    p_lease_generation,p_original_owner,p_effect,p_request,p_unknown,
    p_cleanup_generation,p_cleanup_owner,p_observed) $f$;
CREATE FUNCTION tenant_a.test_analytics_maintenance_unknown_defer(
    p_kind text,p_schema name,p_export uuid,p_group text,p_group_generation bigint,
    p_cutover_generation bigint,p_job uuid,p_attempt uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_owner uuid,p_effect uuid,p_request bytea,
    p_unknown timestamptz,p_cleanup_generation bigint,p_cleanup_owner uuid,
    p_observed timestamptz
) RETURNS timestamptz LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $f$
SELECT platform.platform_async_analytics_dashboard_export_maintenance_unknown_defer_v1(
    p_kind,'11111111-1111-4111-8111-111111111111',p_schema,p_export,p_group,
    p_group_generation,p_cutover_generation,p_job,p_attempt,p_attempt_number,
    p_lease_generation,p_original_owner,p_effect,p_request,p_unknown,
    p_cleanup_generation,p_cleanup_owner,p_observed) $f$;
CREATE FUNCTION tenant_a.test_analytics_maintenance_unknown_resolve(
    p_kind text,p_schema name,p_export uuid,p_group text,p_group_generation bigint,
    p_cutover_generation bigint,p_job uuid,p_attempt uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_owner uuid,p_effect uuid,p_request bytea,
    p_unknown timestamptz,p_cleanup_generation bigint,p_cleanup_owner uuid,
    p_resolution_receipt uuid,p_effect_receipt uuid,p_effect_receipt_sha bytea,
    p_terminal text,p_resolution_sha bytea,p_resolved timestamptz
) RETURNS TABLE(replayed boolean,effect_receipt_id uuid,job_state text)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $f$
SELECT * FROM platform.platform_async_analytics_dashboard_export_maintenance_unknown_resolve_v1(
    p_kind,'11111111-1111-4111-8111-111111111111',p_schema,p_export,p_group,
    p_group_generation,p_cutover_generation,p_job,p_attempt,p_attempt_number,
    p_lease_generation,p_original_owner,p_effect,p_request,p_unknown,
    p_cleanup_generation,p_cleanup_owner,p_resolution_receipt,p_effect_receipt,
    p_effect_receipt_sha,p_terminal,p_resolution_sha,p_resolved) $f$;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA tenant_a TO harpy_runtime;`
