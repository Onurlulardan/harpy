-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_setup_read_v1(
    p_scope_schema name, p_tenant_id uuid, p_binding_id uuid,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    binding_id uuid, user_id uuid, setup_state text, approval_state text,
    result_revision bigint, setup_completed_at timestamptz,
    submitted_at timestamptz, decision_at timestamptz,
    updated_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_tenant_id IS NULL OR p_binding_id IS NULL OR
       NOT EXISTS (
          SELECT 1
          FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id
            AND tenant.schema_name::text IS NOT DISTINCT FROM p_scope_schema::text
       ) OR NOT iam_identity_query_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM portal setup query authority is denied';
    END IF;

    RETURN QUERY
    SELECT setup.binding_id, setup.user_id, setup.setup_state,
        setup.approval_state, setup.revision, setup.setup_completed_at,
        setup.submitted_at, setup.decision_at, setup.updated_at
    FROM iam_portal_access_setups AS setup
    WHERE setup.binding_id = p_binding_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM portal setup was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_setup_read_v1(
    name,uuid,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
