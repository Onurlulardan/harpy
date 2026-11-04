-- +goose Up
CREATE FUNCTION platform.tenancy_customer_setup_reissue_target_v1(
    in_tenant_id uuid
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
    updated_at timestamptz,
    setup_reissue_operation_id uuid,
    schema_name text,
    first_user_id uuid,
    first_admin_email text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_tenant_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        NULL::text,
        NULL::boolean,
        NULL::text,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text
    ;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_customer_setup_reissue_target_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_setup_reissue_target_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_setup_reissue_target_v1(uuid);
