-- +goose Up
CREATE FUNCTION platform.platformauthority_installation_head_pin_v3(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_schema_name name,
    in_registry_binding_id uuid,
    in_scope_kind text,
    in_scope_id uuid,
    in_generation bigint,
    in_previous_head_sha256 text,
    in_state_sha256 text,
    in_proposal_sha256 text,
    in_ack_sha256 text,
    in_manifest_sha256 text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text);
