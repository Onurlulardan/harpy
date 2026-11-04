-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_portal_setup_complete_v2(
    in_schema_name name,
    in_tenant_id uuid,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_selector bytea,
    in_hash_key_id text,
    in_secret_hmac bytea,
    in_binding_sha256 bytea,
    in_salt bytea,
    in_tag bytea
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
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_setup_complete_v2(name, uuid, uuid, bytea, bytea, text, bytea, bytea, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_portal_setup_complete_v2(name, uuid, uuid, bytea, bytea, text, bytea, bytea, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_portal_setup_complete_v2(name, uuid, uuid, bytea, bytea, text, bytea, bytea, bytea, bytea);
