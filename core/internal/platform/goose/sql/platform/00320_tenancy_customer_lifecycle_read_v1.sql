-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_read_v1(
    in_control_id uuid,
    in_operation_id uuid
)
RETURNS TABLE (
    operation_id uuid,
    control_tenant_id uuid,
    target_tenant_id uuid,
    kind text,
    state text,
    result_revision bigint,
    replay boolean,
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
        in_operation_id,
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

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_read_v1(uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_read_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_lifecycle_read_v1(uuid, uuid);
