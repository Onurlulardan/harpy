-- +goose Up
CREATE FUNCTION platform.platformauthority_delivery_await_import_v1(
    in_operation_id text,
    in_lease_owner_id text,
    in_lease_generation bigint
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_delivery_await_import_v1(text, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_delivery_await_import_v1(text, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_delivery_await_import_v1(text, text, bigint);
