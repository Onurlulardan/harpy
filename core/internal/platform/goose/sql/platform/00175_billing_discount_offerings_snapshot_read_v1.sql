-- +goose Up
-- BELIRSIZ: discount offering links are not determined by the caller.
CREATE FUNCTION platform.billing_discount_offerings_snapshot_read_v1(
    p_authority_id uuid,
    p_version_id uuid
)
RETURNS TABLE (
    offering_id uuid
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::uuid
    WHERE false;
$$;

REVOKE ALL ON FUNCTION platform.billing_discount_offerings_snapshot_read_v1(
    uuid, uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_discount_offerings_snapshot_read_v1(
    uuid, uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_discount_offerings_snapshot_read_v1(
    uuid, uuid
);
