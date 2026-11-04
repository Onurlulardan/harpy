-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_instance_owner_input_v1(
    in_environment_id uuid
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    schema_name text,
    user_id uuid,
    owner_binding_id uuid,
    request_sha256 text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        stored.operation_id,
        stored.tenant_id,
        stored.schema_name::text,
        stored.user_id,
        stored.owner_binding_id,
        stored.request_sha256
    FROM platform.iam_installation_instance_owner_inputs AS stored
    WHERE stored.environment_id = in_environment_id
      AND stored.operation_id IS NOT NULL
      AND stored.operation_id <> '00000000-0000-0000-0000-000000000000'::uuid
      AND stored.tenant_id IS NOT NULL
      AND stored.tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid
      AND stored.user_id IS NOT NULL
      AND stored.user_id <> '00000000-0000-0000-0000-000000000000'::uuid;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_input_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_input_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_instance_owner_input_v1(uuid);
