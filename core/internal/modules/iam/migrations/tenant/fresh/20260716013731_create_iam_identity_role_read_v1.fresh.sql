-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_role_read_v1(
    p_scope_schema name, p_role_id uuid, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, scope_kind text, scope_id uuid, organization_id uuid,
    name text, description text, status text, is_protected boolean,
    is_default boolean, revision bigint, created_at timestamptz,
    updated_at timestamptz, retired_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_role_id IS NULL OR NOT iam_identity_query_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY SELECT role.id, role.scope_kind, role.scope_id,
        COALESCE(role.organization_id, '00000000-0000-0000-0000-000000000000'::uuid),
        role.name, COALESCE(role.description, ''), role.status,
        role.is_protected, role.is_default, role.revision,
        role.created_at, role.updated_at, role.retired_at
    FROM iam_roles AS role WHERE role.id = p_role_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_role_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
