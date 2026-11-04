-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1()
RETURNS TABLE (
    tenant_id uuid,
    tenant_schema text,
    export_id uuid,
    job_id uuid,
    attempt_id uuid,
    attempt_generation bigint,
    lease_owner_id uuid,
    effect_started boolean,
    effect_id uuid,
    effect_request_sha256 bytea,
    cancellation_requested boolean,
    lease_expired_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        false,
        NULL::uuid,
        ''::bytea,
        false,
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1();
