-- +goose Up
CREATE FUNCTION platform.tenancy_customer_provision_read_v1(
    in_operation_id uuid,
    in_subject_id uuid
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    requested_by_subject_id uuid,
    state text,
    setup_delivery_state text,
    failure_code text,
    failure_retryable boolean,
    correlation_id text,
    created_at timestamptz,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_operation_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        NULL::text,
        NULL::boolean,
        NULL::text,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp()
    ;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_customer_provision_read_v1(uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_provision_read_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_provision_read_v1(uuid, uuid);
