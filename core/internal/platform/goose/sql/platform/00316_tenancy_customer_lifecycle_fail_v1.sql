-- +goose Up
CREATE FUNCTION platform.tenancy_customer_lifecycle_fail_v1(
    in_id uuid,
    in_worker_id uuid,
    in_code text,
    in_now timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_fail_v1(uuid, uuid, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_fail_v1(uuid, uuid, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_lifecycle_fail_v1(uuid, uuid, text, timestamptz);
