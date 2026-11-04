-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_import_source_cleanup_claim_v1(
    in_schema_name name,
    in_id uuid,
    in_lease_owner_id uuid,
    in_utc timestamptz
)
RETURNS TABLE (
    source_id uuid,
    actor_id uuid,
    resource_kind text,
    object_key text,
    object_version text,
    cleanup_attempt bigint,
    lease_generation bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_system_import_source_cleanup_claim_v1(name, uuid, uuid, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_import_source_cleanup_claim_v1(name, uuid, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_import_source_cleanup_claim_v1(name, uuid, uuid, timestamptz);
