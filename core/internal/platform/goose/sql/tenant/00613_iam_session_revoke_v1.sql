-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_session_revoke_v1(
    in_schema_name name,
    in_token_id uuid,
    in_family_id uuid,
    in_expected_token_revision bigint,
    in_expected_family_revision bigint,
    in_session_id uuid,
    in_expected_session_revision bigint,
    in_user_id uuid,
    in_expected_user_revision bigint
)
RETURNS TABLE (
    outcome text,
    locator_selector bytea,
    locator_expected_revision bigint,
    locator_revision bigint,
    locator_source_sha256 bytea,
    locator_state text,
    locator_projected_at timestamptz,
    locator_retain_until timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::bytea,
        NULL::bigint,
        NULL::bigint,
        ''::bytea,
        NULL::text,
        NULL::timestamptz,
        NULL::timestamptz
    ;
$$;

REVOKE ALL ON FUNCTION iam_session_revoke_v1(name, uuid, uuid, bigint, bigint, uuid, bigint, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_session_revoke_v1(name, uuid, uuid, bigint, bigint, uuid, bigint, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_session_revoke_v1(name, uuid, uuid, bigint, bigint, uuid, bigint, uuid, bigint);
