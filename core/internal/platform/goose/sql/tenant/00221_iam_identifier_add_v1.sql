-- +goose Up
CREATE FUNCTION iam_identifier_add_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_user_id uuid,
    in_kind text,
    in_display_value text,
    in_normalized_value text,
    in_normalized_sha256 bytea,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    subject_user_id uuid,
    revision bigint,
    token_version bigint,
    authorization_version bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_identifier_add_v1(name, uuid, uuid, bytea, uuid, uuid, text, text, text, bytea, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identifier_add_v1(name, uuid, uuid, bytea, uuid, uuid, text, text, text, bytea, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identifier_add_v1(name, uuid, uuid, bytea, uuid, uuid, text, text, text, bytea, uuid, uuid, bigint, bigint);
