-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1()
RETURNS TABLE (
    tenant_id uuid,
    tenant_schema text,
    export_id uuid,
    job_id uuid,
    attempt_generation bigint
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
        0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1();
