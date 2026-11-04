-- +goose Up
CREATE FUNCTION platform.iam_login_directory_upsert_v2(
    in_tenant_id uuid,
    in_group_generation bigint,
    in_generation bigint,
    in_user_id uuid,
    in_identifier_id uuid,
    in_kind text,
    in_lookup_key_id text,
    in_lookup_hmac bytea,
    in_revision bigint,
    in_normalized_sha256 bytea
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.iam_login_directory_upsert_v2(uuid, bigint, bigint, uuid, uuid, text, text, bytea, bigint, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_login_directory_upsert_v2(uuid, bigint, bigint, uuid, uuid, text, text, bytea, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_login_directory_upsert_v2(uuid, bigint, bigint, uuid, uuid, text, text, bytea, bigint, bytea);
