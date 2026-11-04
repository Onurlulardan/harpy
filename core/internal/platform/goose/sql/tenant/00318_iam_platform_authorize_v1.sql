-- +goose Up
CREATE FUNCTION iam_platform_authorize_v1(
    in_schema_name name,
    in_actor_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_expected_platform_clock bigint,
    in_capability text
)
RETURNS TABLE (
    allowed boolean,
    capability_key text,
    deployment_mode text,
    subject_id uuid,
    platform_clock bigint,
    authorization_version bigint,
    reason text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        false,
        ''::text,
        ''::text,
        NULL::uuid,
        0::bigint,
        0::bigint,
        ''::text
    ;
$$;

REVOKE ALL ON FUNCTION iam_platform_authorize_v1(name, uuid, bigint, bigint, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_platform_authorize_v1(name, uuid, bigint, bigint, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_platform_authorize_v1(name, uuid, bigint, bigint, bigint, text);
