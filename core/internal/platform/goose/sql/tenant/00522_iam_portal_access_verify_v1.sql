-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_portal_access_verify_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_principal_id uuid,
    in_binding_id uuid,
    in_token_version bigint,
    in_authorization_version bigint
)
RETURNS TABLE (
    principal_id uuid,
    binding_id uuid,
    tenant_id uuid,
    token_version bigint,
    authorization_version bigint,
    authorization_clock bigint,
    binding_revision bigint,
    subject_type text,
    form_id uuid,
    form_version_id uuid,
    subject_record_id uuid,
    identifier_field_id uuid,
    role_names text[]
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_principal_id,
        in_binding_id,
        in_tenant_id,
        in_token_version,
        in_authorization_version,
        0::bigint,
        0::bigint,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '{}'::text[]
    ;
$$;

REVOKE ALL ON FUNCTION iam_portal_access_verify_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_portal_access_verify_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_portal_access_verify_v1(name, uuid, uuid, uuid, bigint, bigint);
