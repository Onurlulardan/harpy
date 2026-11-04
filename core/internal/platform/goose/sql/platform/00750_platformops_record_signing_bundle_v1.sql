-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_record_signing_bundle_v1(
    in_environment_id text,
    in_generation bigint,
    in_sha256 text,
    in_bundle_bytes bytea,
    in_created_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_record_signing_bundle_v1(text, bigint, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_record_signing_bundle_v1(text, bigint, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_record_signing_bundle_v1(text, bigint, text, bytea, timestamptz);
