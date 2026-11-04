-- +goose Up
CREATE FUNCTION platform.platformauthority_delivery_retry_v1(
    in_operation_id text,
    in_lease_owner_id text,
    in_lease_generation bigint,
    in_error_code text,
    in_permanent boolean
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_delivery_retry_v1(text, text, bigint, text, boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_delivery_retry_v1(text, text, bigint, text, boolean) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_delivery_retry_v1(text, text, bigint, text, boolean);
