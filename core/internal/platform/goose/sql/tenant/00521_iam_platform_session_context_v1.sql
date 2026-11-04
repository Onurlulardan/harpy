-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_platform_session_context_v1(
    in_schema_name name,
    in_actor_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    deployment_mode text,
    tenant_purpose text,
    subject_id bigint,
    platform_clock bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        0::bigint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_platform_session_context_v1(name, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_platform_session_context_v1(name, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_platform_session_context_v1(name, uuid, bigint, bigint);
