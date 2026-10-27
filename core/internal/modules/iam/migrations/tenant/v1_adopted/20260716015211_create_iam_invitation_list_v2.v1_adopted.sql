-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_list_v2(
    p_scope_schema name, p_actor_user_id uuid, p_states jsonb,
    p_after uuid, p_limit integer
)
RETURNS TABLE (
    invitation_id uuid, display_value text, normalized_value text,
    role_id uuid, role_name text, role_scope_kind text, role_scope_id uuid,
    organization_id uuid, invited_by uuid, accepted_user_id uuid,
    accepted_identifier_id uuid, invitation_state text,
    invitation_revision bigint, created_at timestamptz,
    updated_at timestamptz, expires_at timestamptz, terminal_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN QUERY
    SELECT legacy.invitation_id, legacy.display_value, legacy.normalized_value,
           legacy.role_id, role.name, legacy.role_scope_kind, legacy.role_scope_id,
           legacy.organization_id, legacy.invited_by, legacy.accepted_user_id,
           legacy.accepted_identifier_id, legacy.invitation_state,
           legacy.invitation_revision, legacy.created_at, legacy.updated_at,
           legacy.expires_at, legacy.terminal_at
    FROM iam_invitation_list_v1(
        p_scope_schema, p_actor_user_id, p_states, p_after, p_limit
    ) AS legacy
    JOIN iam_roles AS role ON role.id = legacy.role_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_list_v2(
    name,uuid,jsonb,uuid,integer
) FROM PUBLIC;

