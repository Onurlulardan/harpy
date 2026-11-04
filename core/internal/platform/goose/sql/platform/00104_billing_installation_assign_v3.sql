-- +goose Up
CREATE FUNCTION platform.billing_installation_assign_v3(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_authority_id uuid,
    in_assignment_id uuid,
    in_deployment_mode text,
    in_offering_version_id uuid,
    in_offering_revision bigint,
    in_offering_sha256 text,
    in_offering_effective_at timestamptz,
    in_source_sequence bigint,
    in_source_sha256 text,
    in_manifest_sha256 text,
    in_binding_purpose text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text);
