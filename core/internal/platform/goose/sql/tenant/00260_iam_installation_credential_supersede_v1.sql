-- +goose Up
CREATE FUNCTION iam_installation_credential_supersede_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_tenant_id uuid,
    in_first_owner_user_id uuid,
    in_current_credential_id uuid,
    in_current_credential_version bigint,
    in_next_credential_id uuid,
    in_next_credential_version bigint,
    in_salt bytea,
    in_tag bytea
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    user_id uuid,
    previous_credential_id uuid,
    credential_id uuid,
    credential_version bigint,
    credential_revision bigint,
    previous_marker_revision bigint,
    marker_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_operation_id,
        in_tenant_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_installation_credential_supersede_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, bigint, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_installation_credential_supersede_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, bigint, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_installation_credential_supersede_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, bigint, bytea, bytea);
