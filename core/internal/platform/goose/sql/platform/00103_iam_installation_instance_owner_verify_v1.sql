-- +goose Up
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
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_verify_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_verify_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_instance_owner_verify_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text);
