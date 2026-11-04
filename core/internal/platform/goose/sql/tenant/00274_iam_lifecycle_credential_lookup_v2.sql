-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_lifecycle_credential_lookup_v2(
    in_schema_name name,
    in_purpose text,
    in_selector bytea,
    in_binding_sha256 bytea
)
RETURNS TABLE (
    token_id uuid,
    purpose text,
    user_id uuid,
    identifier_id uuid,
    identifier_kind text,
    identifier_revision bigint,
    portal_binding_id uuid,
    portal_revision bigint,
    selector bytea,
    hash_key_id text,
    secret_hmac bytea,
    binding bytea,
    revision bigint,
    state text,
    expires_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_purpose,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        0::bigint,
        NULL::uuid,
        NULL::bigint,
        in_selector,
        ''::text,
        ''::bytea,
        ''::bytea,
        0::bigint,
        ''::text,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_lifecycle_credential_lookup_v2(name, text, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_lifecycle_credential_lookup_v2(name, text, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_lifecycle_credential_lookup_v2(name, text, bytea, bytea);
