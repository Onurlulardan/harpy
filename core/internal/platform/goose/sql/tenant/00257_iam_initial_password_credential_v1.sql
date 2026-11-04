-- +goose Up
CREATE FUNCTION iam_initial_password_credential_v1(
    in_schema_name name,
    in_source_user_id uuid,
    in_source_identifier_id uuid,
    in_source_identifier_kind text,
    in_source_revision bigint,
    in_binding bytea
)
RETURNS TABLE (
    identifier_id uuid,
    user_id uuid,
    user_revision bigint,
    token_version bigint,
    authorization_version bigint,
    access_mode_id uuid,
    access_mode_revision bigint,
    installation_operation_id uuid,
    credential_id uuid,
    credential_version bigint,
    credential_revision bigint,
    algorithm text,
    salt bytea,
    tag bytea,
    memory_kib integer,
    iterations smallint,
    parallelism smallint,
    credential_created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        ''::text,
        ''::bytea,
        ''::bytea,
        0::integer,
        0::smallint,
        0::smallint,
        pg_catalog.clock_timestamp()
    ;
$$;

REVOKE ALL ON FUNCTION iam_initial_password_credential_v1(name, uuid, uuid, text, bigint, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_initial_password_credential_v1(name, uuid, uuid, text, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_initial_password_credential_v1(name, uuid, uuid, text, bigint, bytea);
