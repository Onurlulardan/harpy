-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_login_credential_v1(
    in_schema_name name,
    in_source_user_id uuid,
    in_source_identifier_id uuid,
    in_source_identifier_kind text,
    in_source_revision bigint,
    in_binding bytea,
    in_mode text
)
RETURNS TABLE (
    identifier_id uuid,
    user_id uuid,
    user_revision bigint,
    token_version bigint,
    authorization_version bigint,
    access_mode_id uuid,
    access_mode_revision bigint,
    portal_binding_id uuid,
    portal_binding_revision bigint,
    credential_version bigint,
    credential_revision bigint,
    algorithm text,
    salt bytea,
    tag bytea,
    memory_kib integer,
    iterations smallint,
    parallelism smallint,
    legacy_bcrypt_hash text,
    credential_created_at timestamptz
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
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        NULL::uuid,
        NULL::bigint,
        0::bigint,
        0::bigint,
        ''::text,
        ''::bytea,
        ''::bytea,
        NULL::integer,
        NULL::smallint,
        NULL::smallint,
        NULL::text,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_login_credential_v1(name, uuid, uuid, text, bigint, bytea, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_login_credential_v1(name, uuid, uuid, text, bigint, bytea, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_login_credential_v1(name, uuid, uuid, text, bigint, bytea, text);
