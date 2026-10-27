-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_read_v1(
    p_scope_schema name,
    p_actor_user_id uuid,
    p_invitation_id uuid
)
RETURNS TABLE (
    invitation_id uuid, display_value text, normalized_value text,
    role_id uuid, role_scope_kind text, role_scope_id uuid,
    organization_id uuid, invited_by uuid, accepted_user_id uuid,
    accepted_identifier_id uuid, invitation_state text,
    invitation_revision bigint, created_at timestamptz,
    updated_at timestamptz, expires_at timestamptz, terminal_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation read scope is invalid';
    END IF;
    IF p_actor_user_id IS NULL OR p_invitation_id IS NULL OR NOT EXISTS (
        SELECT 1 FROM iam_users AS actor
        WHERE actor.id OPERATOR(pg_catalog.=) p_actor_user_id
          AND actor.status OPERATOR(pg_catalog.=) 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation read authority is denied';
    END IF;
    RETURN QUERY
    SELECT invitation.id, invitation.display_value, invitation.normalized_value,
           invitation.role_id, invitation.role_scope_kind,
           invitation.role_scope_id, invitation.organization_id,
           invitation.invited_by, invitation.accepted_user_id,
           invitation.accepted_identifier_id, invitation.state,
           invitation.revision, invitation.created_at, invitation.updated_at,
           invitation.expires_at, COALESCE(
               invitation.accepted_at,
               invitation.revoked_at,
               invitation.expired_at
           )
    FROM iam_invitations AS invitation
    WHERE invitation.id OPERATOR(pg_catalog.=) p_invitation_id AND (
        invitation.invited_by OPERATOR(pg_catalog.=) p_actor_user_id OR
        iam_invitation_manager_v1(p_actor_user_id)
    );
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM invitation is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_read_v1(name,uuid,uuid) FROM PUBLIC;
