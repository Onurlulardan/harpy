-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_record_v2(
    in_operation_id uuid,
    in_tenant_id uuid,
    in_user_revision_id uuid,
    in_organization_revision_id uuid,
    in_role_revision_id uuid,
    in_initialized_at timestamptz,
    in_schema_name name
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT '\x0000000000000000000000000000000000000000000000000000000000000000'::bytea;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name);
