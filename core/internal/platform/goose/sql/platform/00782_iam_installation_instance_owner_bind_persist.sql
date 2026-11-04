-- +goose Up
DROP FUNCTION platform.iam_installation_instance_owner_bind_v2(uuid, uuid, uuid, name, uuid, uuid, uuid, uuid, uuid, bigint, text);

-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_instance_owner_bind_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_schema_name name,
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
    WITH accepted AS (
        SELECT
            in_operation_id AS operation_id,
            in_environment_id AS environment_id,
            in_tenant_id AS tenant_id,
            in_schema_name AS schema_name,
            in_user_id AS user_id,
            in_subject_id AS subject_id,
            in_owner_role_id AS owner_role_id,
            in_role_permission_id AS role_permission_id,
            in_owner_binding_id AS owner_binding_id,
            in_authorization_version AS authorization_version,
            in_request_sha256 AS request_sha256
        WHERE in_operation_id IS NOT NULL
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
          AND pg_catalog.length(in_schema_name::text) > 0
          AND pg_catalog.length(in_request_sha256) > 0
    ), inserted AS (
        INSERT INTO platform.iam_installation_instance_owner_inputs (
            operation_id, environment_id, tenant_id, schema_name, user_id,
            subject_id, owner_role_id, role_permission_id, owner_binding_id,
            authorization_version, request_sha256
        )
        SELECT
            accepted.operation_id, accepted.environment_id, accepted.tenant_id,
            accepted.schema_name, accepted.user_id, accepted.subject_id,
            accepted.owner_role_id, accepted.role_permission_id,
            accepted.owner_binding_id, accepted.authorization_version,
            accepted.request_sha256
        FROM accepted
        ON CONFLICT (environment_id) DO NOTHING
        RETURNING environment_id
    )
    SELECT CASE
        WHEN EXISTS (SELECT 1 FROM inserted) THEN false
        WHEN EXISTS (SELECT 1 FROM accepted) AND EXISTS (
            SELECT 1
            FROM platform.iam_installation_instance_owner_inputs AS stored
            WHERE stored.environment_id = in_environment_id
        ) THEN true
        ELSE false
    END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_bind_v2(uuid, uuid, uuid, name, uuid, uuid, uuid, uuid, uuid, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_bind_v2(uuid, uuid, uuid, name, uuid, uuid, uuid, uuid, uuid, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_instance_owner_bind_v2(uuid, uuid, uuid, name, uuid, uuid, uuid, uuid, uuid, bigint, text);

-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_instance_owner_bind_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_schema_name name,
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
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_bind_v2(uuid, uuid, uuid, name, uuid, uuid, uuid, uuid, uuid, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_bind_v2(uuid, uuid, uuid, name, uuid, uuid, uuid, uuid, uuid, bigint, text) TO harpy_api, harpy_worker;
