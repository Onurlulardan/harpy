-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_portal_user_status_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_user_id uuid,
    in_expected_revision bigint,
    in_status text,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    revision bigint,
    authorization_clock bigint,
    replayed boolean,
    email_identifier_id uuid,
    email_revision bigint,
    portal_identifier_id uuid,
    portal_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        false,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_user_status_v1(name, uuid, uuid, bytea, uuid, bigint, text, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_portal_user_status_v1(name, uuid, uuid, bytea, uuid, bigint, text, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_portal_user_status_v1(name, uuid, uuid, bytea, uuid, bigint, text, uuid, uuid, bigint, bigint);
