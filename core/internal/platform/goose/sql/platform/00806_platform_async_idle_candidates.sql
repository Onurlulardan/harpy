-- +goose Up
DROP FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1()
RETURNS TABLE (
    tenant_id uuid, tenant_schema text, export_id uuid, job_id uuid,
    attempt_generation bigint
)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::bigint
    WHERE false;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1() TO harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1()
RETURNS TABLE (
    tenant_id uuid, tenant_schema text, export_id uuid, job_id uuid,
    attempt_id uuid, attempt_generation bigint, lease_owner_id uuid,
    effect_started boolean, effect_id uuid, effect_request_sha256 bytea,
    cancellation_requested boolean, lease_expired_at timestamptz
)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::uuid,
        NULL::bigint, NULL::uuid, NULL::boolean, NULL::uuid, NULL::bytea,
        NULL::boolean, NULL::timestamptz
    WHERE false;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1() TO harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1()
RETURNS TABLE (
    tenant_id uuid, tenant_schema text, export_id uuid, job_id uuid,
    attempt_id uuid, attempt_generation bigint, lease_owner_id uuid,
    effect_id uuid, effect_request_sha256 bytea, unknown_at timestamptz
)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::uuid, NULL::text, NULL::uuid, NULL::uuid, NULL::uuid,
        NULL::bigint, NULL::uuid, NULL::uuid, NULL::bytea, NULL::timestamptz
    WHERE false;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1() TO harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1()
RETURNS TABLE (
    kind text, tenant_id uuid, tenant_schema text, export_id uuid,
    authority_group_id text, authority_group_generation bigint,
    cutover_generation bigint, job_id uuid, attempt_id uuid,
    attempt_number bigint, lease_generation bigint,
    original_lease_owner_id uuid, effect_id uuid,
    effect_request_sha256 bytea, unknown_at timestamptz
)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::text, NULL::uuid, NULL::text, NULL::uuid, NULL::text,
        NULL::bigint, NULL::bigint, NULL::uuid, NULL::uuid, NULL::bigint,
        NULL::bigint, NULL::uuid, NULL::uuid, NULL::bytea, NULL::timestamptz
    WHERE false;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1() TO harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid);
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(
    in_after_value uuid
)
RETURNS TABLE (tenant_id uuid, tenant_schema text)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::uuid, NULL::text
    WHERE false AND (in_after_value IS NULL OR in_after_value IS NOT NULL);
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1()
RETURNS TABLE (tenant_id uuid, tenant_schema text, export_id uuid, job_id uuid, attempt_generation bigint)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid, ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid, 0::bigint;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1() TO harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid);
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(in_after_value uuid)
RETURNS TABLE (tenant_id uuid, tenant_schematext text)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid, ''::text
    WHERE in_after_value IS NULL OR in_after_value IS NOT NULL;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid) TO harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1()
RETURNS TABLE (
    tenant_id uuid, tenant_schema text, export_id uuid, job_id uuid, attempt_id uuid,
    attempt_generation bigint, lease_owner_id uuid, effect_started boolean,
    effect_id uuid, effect_request_sha256 bytea, cancellation_requested boolean,
    lease_expired_at timestamptz
)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid, ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid, 0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid, false, NULL::uuid,
        ''::bytea, false, pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1() TO harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1()
RETURNS TABLE (
    tenant_id uuid, tenant_schema text, export_id uuid, job_id uuid, attempt_id uuid,
    attempt_generation bigint, lease_owner_id uuid, effect_id uuid,
    effect_request_sha256 bytea, unknown_at timestamptz
)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid, ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid, 0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid, ''::bytea,
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1() TO harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1()
RETURNS TABLE (
    kind text, tenant_id uuid, tenant_schema text, export_id uuid,
    authority_group_id text, authority_group_generation bigint, cutover_generation bigint,
    job_id uuid, attempt_id uuid, attempt_number bigint, lease_generation bigint,
    original_lease_owner_id uuid, effect_id uuid, effect_request_sha256 bytea,
    unknown_at timestamptz
)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT ''::text, '00000000-0000-0000-0000-000000000000'::uuid, ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid, ''::text, 0::bigint, 0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid, 0::bigint, 0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid, ''::bytea,
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1() TO harpy_api, harpy_worker;
