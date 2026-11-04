-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_initialize_v4(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_user_id uuid,
    in_organization_id uuid,
    in_protected_assignment_id uuid,
    in_email text,
    in_display_name text,
    in_tenant_name text,
    in_binding_set_sha256 text,
    in_user_schema_id uuid,
    in_user_field_id uuid,
    in_organization_schema_id uuid,
    in_organization_field_id uuid,
    in_role_schema_id uuid,
    in_role_field_id uuid,
    in_identifier_id uuid,
    in_access_mode_id uuid,
    in_membership_id uuid,
    in_role_id uuid,
    in_permission_id uuid,
    in_role_permission_id uuid
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

REVOKE ALL ON FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid);
