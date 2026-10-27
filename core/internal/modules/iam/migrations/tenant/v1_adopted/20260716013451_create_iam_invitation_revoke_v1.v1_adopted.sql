-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_revoke_v1(
    p_scope_schema name,
    p_invitation_id uuid,
    p_expected_revision bigint,
    p_actor_user_id uuid,
    p_expected_authorization_version bigint
)
RETURNS TABLE (invitation_id uuid, invitation_state text, invitation_revision bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_state text;
    v_revision bigint;
    v_expires_at timestamptz;
    v_invited_by uuid;
    v_actor_state text;
    v_authorization_version bigint;
    v_token_count integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation revocation scope is invalid';
    END IF;
    IF p_invitation_id IS NULL OR p_expected_revision < 1 OR
       p_actor_user_id IS NULL OR p_expected_authorization_version < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM invitation revocation input is invalid';
    END IF;
    SELECT invitation.state, invitation.revision,
           invitation.expires_at, invitation.invited_by,
           actor.status, actor.authorization_version
    INTO v_state, v_revision, v_expires_at, v_invited_by,
         v_actor_state, v_authorization_version
    FROM iam_invitations AS invitation
    JOIN iam_users AS actor ON actor.id = p_actor_user_id
    WHERE invitation.id OPERATOR(pg_catalog.=) p_invitation_id
    FOR UPDATE OF invitation, actor;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM invitation is unknown';
    END IF;
    IF v_state IS DISTINCT FROM 'pending' OR
       v_revision IS DISTINCT FROM p_expected_revision OR
       v_now >= v_expires_at OR v_actor_state IS DISTINCT FROM 'active' OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM invitation revocation CAS is stale';
    END IF;
    IF v_invited_by IS DISTINCT FROM p_actor_user_id AND
       NOT iam_invitation_manager_v1(p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation revocation authority is denied';
    END IF;
    UPDATE iam_one_time_tokens AS token
    SET state = 'revoked', revoked_at = v_now, updated_at = v_now,
        revision = revision + 1
    WHERE token.invitation_id = p_invitation_id
      AND token.purpose = 'invitation_accept_v1' AND token.state = 'pending';
    GET DIAGNOSTICS v_token_count = ROW_COUNT;
    IF v_token_count IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM invitation token revocation CAS is stale';
    END IF;
    UPDATE iam_invitations
    SET state = 'revoked', revoked_at = v_now, updated_at = v_now,
        revision = v_revision + 1
    WHERE id = p_invitation_id;
    RETURN QUERY SELECT p_invitation_id, 'revoked'::text, v_revision + 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_revoke_v1(
    name,uuid,bigint,uuid,bigint
) FROM PUBLIC;
