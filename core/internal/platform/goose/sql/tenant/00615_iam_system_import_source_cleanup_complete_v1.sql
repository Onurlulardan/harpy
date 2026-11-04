-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_import_source_cleanup_complete_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_source_id uuid,
    in_actor_id uuid,
    in_resource text,
    in_object_key text,
    in_object_version text,
    in_lease_owner_id uuid,
    in_lease_generation bigint,
    in_provider_request_id text,
    in_delete_response_sha256 bytea,
    in_absence_probe_sha256 bytea,
    in_utc timestamptz
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

REVOKE ALL ON FUNCTION iam_system_import_source_cleanup_complete_v1(name, uuid, uuid, uuid, text, text, text, uuid, bigint, text, bytea, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_import_source_cleanup_complete_v1(name, uuid, uuid, uuid, text, text, text, uuid, bigint, text, bytea, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_import_source_cleanup_complete_v1(name, uuid, uuid, uuid, text, text, text, uuid, bigint, text, bytea, bytea, timestamptz);
