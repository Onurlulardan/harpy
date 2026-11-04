-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_identity_directory_upsert_v1(
    in_tenant_id uuid,
    in_group_generation bigint,
    in_generation bigint,
    in_id uuid,
    in_identifier_id uuid,
    in_lookup_key_id text,
    in_directory_hmac bytea,
    in_revision bigint,
    in_email_sha bytea
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT 0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_identity_directory_upsert_v1(uuid, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_identity_directory_upsert_v1(uuid, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_identity_directory_upsert_v1(uuid, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea);
