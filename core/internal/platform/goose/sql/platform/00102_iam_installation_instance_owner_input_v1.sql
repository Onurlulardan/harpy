-- +goose Up
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
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text;
$$;

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_input_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_input_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_instance_owner_input_v1(uuid);
