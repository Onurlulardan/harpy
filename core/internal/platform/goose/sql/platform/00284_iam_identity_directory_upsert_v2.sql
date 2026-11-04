-- +goose Up
CREATE FUNCTION platform.iam_identity_directory_upsert_v2(
    in_tenant_id uuid,
    in_authority_group_generation bigint,
    in_iam_generation bigint,
    in_accepted_user_id uuid,
    in_accepted_identifier_id uuid,
    in_lookup_key_id text,
    in_hmac bytea,
    in_int64 bigint,
    in_source_sha256 bytea
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.iam_identity_directory_upsert_v2(uuid, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_identity_directory_upsert_v2(uuid, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_identity_directory_upsert_v2(uuid, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea);
