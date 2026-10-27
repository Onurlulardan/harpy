-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_credential_lookup_v2(
    p_scope_schema name, p_selector bytea, p_binding_sha256 bytea
)
RETURNS TABLE (
    invitation_id uuid, display_value text, normalized_value text,
    role_id uuid, role_name text, role_scope_kind text, role_scope_id uuid,
    organization_id uuid, invited_by uuid, accepted_user_id uuid,
    accepted_identifier_id uuid, invitation_state text,
    invitation_revision bigint, created_at timestamptz,
    updated_at timestamptz, expires_at timestamptz, terminal_at timestamptz,
    token_id uuid, selector bytea, hash_key_id text, secret_hmac bytea,
    binding_sha256 bytea, token_revision bigint, profile_catalog jsonb
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_selector) <> 16 OR
       pg_catalog.octet_length(p_binding_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation credential lookup is denied';
    END IF;
    RETURN QUERY
    SELECT invitation.id, invitation.display_value, invitation.normalized_value,
           invitation.role_id, role.name, invitation.role_scope_kind,
           invitation.role_scope_id, invitation.organization_id,
           invitation.invited_by, invitation.accepted_user_id,
           invitation.accepted_identifier_id, invitation.state,
           invitation.revision, invitation.created_at, invitation.updated_at,
           invitation.expires_at, COALESCE(
               invitation.accepted_at, invitation.revoked_at, invitation.expired_at
           ), token.id, token.selector, token.hash_key_id,
           token.secret_hmac, token.binding_sha256, token.revision,
           iam_profile_write_contract_read_v1(p_scope_schema, 'user')
    FROM iam_one_time_tokens AS token
    JOIN iam_invitations AS invitation ON invitation.id = token.invitation_id
    JOIN iam_roles AS role ON role.id = invitation.role_id
    WHERE token.purpose = 'invitation_accept_v1'
      AND token.selector = p_selector
      AND token.binding_sha256 = p_binding_sha256
      AND token.state IN ('pending', 'consumed')
      AND invitation.state IN ('pending', 'accepted');
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM invitation credential is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_credential_lookup_v2(
    name,bytea,bytea
) FROM PUBLIC;
