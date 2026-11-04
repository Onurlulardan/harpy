-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_import_source_cleanup_retry_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_source_id uuid,
    in_lease_owner_id uuid,
    in_lease_generation bigint,
    in_error_code text,
    in_utc timestamptz,
    in_utc_2 timestamptz
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

REVOKE ALL ON FUNCTION iam_system_import_source_cleanup_retry_v1(name, uuid, uuid, uuid, bigint, text, timestamptz, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_import_source_cleanup_retry_v1(name, uuid, uuid, uuid, bigint, text, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_import_source_cleanup_retry_v1(name, uuid, uuid, uuid, bigint, text, timestamptz, timestamptz);
