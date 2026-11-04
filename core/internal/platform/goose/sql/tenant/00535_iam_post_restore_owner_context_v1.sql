-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_post_restore_owner_context_v1(
    in_owner_user_id uuid
)
RETURNS TABLE (
    identifier_id uuid,
    normalized_value text,
    normalized_sha256 bytea,
    revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::bytea,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_post_restore_owner_context_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_post_restore_owner_context_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_post_restore_owner_context_v1(uuid);
