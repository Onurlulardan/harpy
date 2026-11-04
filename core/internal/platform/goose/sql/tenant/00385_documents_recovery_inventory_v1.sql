-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION documents_recovery_inventory_v1(
    arg_1 uuid,
    arg_2 uuid
)
RETURNS TABLE (
    version_id uuid,
    object_key text,
    object_version text,
    object_generation bigint,
    ciphertext_sha256 bytea,
    ciphertext_bytes bigint,
    key_id text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        NULL::bytea,
        0::bigint,
        ''::text;
$$;

REVOKE ALL ON FUNCTION documents_recovery_inventory_v1(uuid, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_inventory_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION documents_recovery_inventory_v1(uuid, uuid);
