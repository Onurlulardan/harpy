-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_capability_grant_write_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_grant_id uuid,
    in_principal_id uuid,
    in_expected_revision bigint,
    in_capability text,
    in_kind text,
    in_owner text,
    in_resource_kind text,
    in_resource_id uuid,
    in_valid_from timestamptz,
    in_valid_until timestamptz,
    in_manifest_sha256 bytea,
    in_actor_id uuid,
    in_manage_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_purpose_sha256 bytea,
    in_auth_time timestamptz,
    in_amr_password boolean,
    in_amr_web_authn boolean,
    in_amrtotp boolean
)
RETURNS TABLE (
    result_id uuid,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
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
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_capability_grant_write_v1(name, uuid, bytea, uuid, uuid, bigint, text, text, text, text, uuid, timestamptz, timestamptz, bytea, uuid, uuid, bigint, bigint, bytea, timestamptz, boolean, boolean, boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_capability_grant_write_v1(name, uuid, bytea, uuid, uuid, bigint, text, text, text, text, uuid, timestamptz, timestamptz, bytea, uuid, uuid, bigint, bigint, bytea, timestamptz, boolean, boolean, boolean) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_capability_grant_write_v1(name, uuid, bytea, uuid, uuid, bigint, text, text, text, text, uuid, timestamptz, timestamptz, bytea, uuid, uuid, bigint, bigint, bytea, timestamptz, boolean, boolean, boolean);
