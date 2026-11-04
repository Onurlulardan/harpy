-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_credential_lookup_v2(
    in_schema_name name,
    in_selector bytea,
    in_binding_sha256 bytea
)
RETURNS TABLE (
    id uuid,
    display_value text,
    normalized_value text,
    role_id uuid,
    role_name text,
    scope_kind text,
    id_2 uuid,
    organization_id uuid,
    invited_by uuid,
    accepted_user_id uuid,
    accepted_identifier_id uuid,
    state text,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    expires_at timestamptz,
    terminal_at timestamptz,
    id_3 uuid,
    selector bytea,
    hash_key_id text,
    secret_hmac bytea,
    binding bytea,
    revision_2 bigint,
    profile_catalog bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        NULL::uuid,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_selector,
        ''::text,
        NULL::bytea,
        NULL::bytea,
        0::bigint,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_credential_lookup_v2(name, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_credential_lookup_v2(name, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_credential_lookup_v2(name, bytea, bytea);
