-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_recovery_credential_v1(
    in_schema_name name,
    in_binding bytea,
    in_hmac_key_id text,
    in_code_hmac bytea
)
RETURNS TABLE (
    code_id uuid,
    authenticator_id uuid,
    user_id uuid,
    code_revision bigint,
    authenticator_revision bigint,
    user_revision bigint,
    token_version bigint,
    code_state text,
    authenticator_state text,
    user_status text,
    identity_kind text,
    hmac_key_id text,
    code_hmac bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        in_hmac_key_id,
        in_code_hmac
    ;
$$;

REVOKE ALL ON FUNCTION iam_recovery_credential_v1(name, bytea, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_recovery_credential_v1(name, bytea, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_recovery_credential_v1(name, bytea, text, bytea);
