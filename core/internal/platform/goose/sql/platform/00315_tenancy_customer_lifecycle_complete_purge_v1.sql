-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_complete_purge_v1(
    in_id uuid,
    in_worker_id uuid,
    in_id_2 uuid,
    in_sha256 bytea,
    in_now timestamptz
)
RETURNS TABLE (
    operation_id uuid,
    control_tenant_id uuid,
    target_tenant_id uuid,
    kind text,
    state text,
    result_revision bigint,
    "false" boolean,
    created_at timestamptz,
    updated_at timestamptz,
    completed_at timestamptz,
    failure_code text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        false,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        NULL::text
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_complete_purge_v1(uuid, uuid, uuid, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_complete_purge_v1(uuid, uuid, uuid, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_lifecycle_complete_purge_v1(uuid, uuid, uuid, bytea, timestamptz);
