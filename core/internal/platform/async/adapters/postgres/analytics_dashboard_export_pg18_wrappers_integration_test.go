//go:build integration

package postgres

const analyticsDashboardExportWorkerWrappersPG18 = `
CREATE FUNCTION tenant_a.test_analytics_export_claim(
    p_export uuid,p_job uuid,p_generation bigint,p_lease uuid
) RETURNS TABLE(claim_state text,attempt_id uuid,claimed_at timestamptz,
    lease_until timestamptz)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT * FROM platform.platform_async_analytics_dashboard_export_claim_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_generation,p_lease) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_checkpoint(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_observed timestamptz
) RETURNS TABLE(lease_until timestamptz,cancellation_requested boolean)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT * FROM platform.platform_async_analytics_dashboard_export_checkpoint_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_observed) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_effect_start(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea,p_started timestamptz
) RETURNS timestamptz LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT platform.platform_async_analytics_dashboard_export_effect_start_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request,p_started) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_succeed(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea,p_receipt uuid,p_result bytea,p_reference text,
    p_bytes bigint,p_completed timestamptz
) RETURNS boolean LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT platform.platform_async_analytics_dashboard_export_succeed_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request,p_receipt,p_result,p_reference,
    p_bytes,p_completed) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_finish(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_class text,p_code text,p_digest bytea,p_retryable boolean,p_reference text,
    p_finished timestamptz
) RETURNS TABLE(job_state text,effect_receipt_id uuid)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT * FROM platform.platform_async_analytics_dashboard_export_finish_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_class,p_code,p_digest,p_retryable,p_reference,p_finished) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_unknown(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea,p_observed timestamptz
) RETURNS boolean LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT platform.platform_async_analytics_dashboard_export_unknown_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request,p_observed) $f$;`

const analyticsDashboardExportRecoveryWrappersPG18 = `
CREATE FUNCTION tenant_a.test_analytics_export_cleanup_claim(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea,p_cleanup_owner uuid,p_observed timestamptz
) RETURNS TABLE(claim_state text,cleanup_generation bigint,lease_until timestamptz)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT * FROM platform.platform_async_analytics_dashboard_export_unknown_cleanup_claim_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request,p_cleanup_owner,p_observed) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_cleanup_checkpoint(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea,p_cleanup_generation bigint,
    p_cleanup_owner uuid,p_observed timestamptz
) RETURNS TABLE(lease_until timestamptz,cancellation_requested boolean)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT * FROM platform.platform_async_analytics_dashboard_export_unknown_cleanup_checkpoint_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request,p_cleanup_generation,
    p_cleanup_owner,p_observed) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_cleanup_defer(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea,p_cleanup_generation bigint,
    p_cleanup_owner uuid,p_observed timestamptz
) RETURNS timestamptz LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT platform.platform_async_analytics_dashboard_export_unknown_cleanup_defer_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request,p_cleanup_generation,
    p_cleanup_owner,p_observed) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_unknown_resolve(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea,p_cleanup_generation bigint,
    p_cleanup_owner uuid,p_receipt uuid,p_kind text,p_digest bytea,
    p_resolved timestamptz
) RETURNS boolean LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT platform.platform_async_analytics_dashboard_export_unknown_resolve_v2(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request,p_cleanup_generation,
    p_cleanup_owner,p_receipt,p_kind,p_digest,p_resolved) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_unknown_evidence(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea
) RETURNS timestamptz LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT platform.platform_async_analytics_dashboard_export_unknown_evidence_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_expired_finalize(
    p_export uuid,p_job uuid,p_attempt uuid,p_generation bigint,p_lease uuid,
    p_effect uuid,p_request bytea,p_observed timestamptz
) RETURNS TABLE(final_state text,finalized_at timestamptz,failure_sha256 bytea)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT * FROM platform.platform_async_analytics_dashboard_export_expired_finalize_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_attempt,
    p_generation,p_lease,p_effect,p_request,p_observed) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_control(
    p_export uuid,p_job uuid,p_requested timestamptz
) RETURNS TABLE(control_state text,changed_at timestamptz)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT * FROM platform.platform_async_analytics_dashboard_export_control_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_requested) $f$;
CREATE FUNCTION tenant_a.test_analytics_export_spool_state(
    p_export uuid,p_generation bigint,p_lease uuid
) RETURNS text LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT platform.platform_async_analytics_dashboard_export_spool_state_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_generation,p_lease) $f$;
GRANT USAGE ON SCHEMA tenant_a TO harpy_runtime;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA tenant_a TO harpy_runtime;`
