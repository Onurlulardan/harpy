-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_instance_owner_verify_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_user_id uuid,
    in_subject_id uuid,
    in_owner_role_id uuid,
    in_role_permission_id uuid,
    in_owner_binding_id uuid,
    in_authorization_version bigint,
    in_request_sha256 text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT EXISTS (
        SELECT 1
        FROM platform.iam_installation_instance_owner_inputs AS stored
        WHERE stored.operation_id = in_operation_id
          AND stored.environment_id = in_environment_id
          AND stored.tenant_id = in_tenant_id
          AND stored.user_id = in_user_id
          AND stored.subject_id = in_subject_id
          AND stored.owner_role_id = in_owner_role_id
          AND stored.role_permission_id = in_role_permission_id
          AND stored.owner_binding_id = in_owner_binding_id
          AND stored.authorization_version = in_authorization_version
          AND stored.request_sha256 = in_request_sha256
          AND in_operation_id IS NOT NULL
          AND in_operation_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_environment_id IS NOT NULL
          AND in_environment_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_tenant_id IS NOT NULL
          AND in_tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_user_id IS NOT NULL
          AND in_user_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_subject_id IS NOT NULL
          AND in_subject_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_owner_role_id IS NOT NULL
          AND in_owner_role_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_role_permission_id IS NOT NULL
          AND in_role_permission_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_owner_binding_id IS NOT NULL
          AND in_owner_binding_id <> '00000000-0000-0000-0000-000000000000'::uuid
    );
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_verify_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_verify_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_instance_owner_verify_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text);
