-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_accept_v1(
    p_scope_schema name,
    p_invitation_id uuid,
    p_expected_invitation_revision bigint,
    p_token_id uuid,
    p_expected_token_revision bigint,
    p_accepted_user_id uuid,
    p_accepted_identifier_id uuid
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_invitation_state text;
    v_invitation_revision bigint;
    v_invitation_expires_at timestamptz;
    v_token_state text;
    v_token_revision bigint;
    v_token_invitation_id uuid;
    v_token_expires_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation acceptance scope is invalid';
    END IF;
    SELECT invitation.state, invitation.revision, invitation.expires_at
    INTO v_invitation_state, v_invitation_revision, v_invitation_expires_at
    FROM iam_invitations AS invitation
    WHERE invitation.id OPERATOR(pg_catalog.=) p_invitation_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM invitation is unknown';
    END IF;
    SELECT token.state, token.revision, token.invitation_id, token.expires_at
    INTO v_token_state, v_token_revision,
         v_token_invitation_id, v_token_expires_at
    FROM iam_one_time_tokens AS token
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
      AND token.purpose OPERATOR(pg_catalog.=) 'invitation_accept_v1'
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM invitation token is unknown';
    END IF;
    IF v_invitation_state IS DISTINCT FROM 'pending' OR
       v_invitation_revision IS DISTINCT FROM p_expected_invitation_revision OR
       v_token_state IS DISTINCT FROM 'pending' OR
       v_token_revision IS DISTINCT FROM p_expected_token_revision OR
       v_token_invitation_id IS DISTINCT FROM p_invitation_id OR
       v_token_expires_at IS DISTINCT FROM v_invitation_expires_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM invitation acceptance CAS is stale';
    END IF;
    IF v_now OPERATOR(pg_catalog.>=) v_invitation_expires_at THEN
        UPDATE iam_one_time_tokens
        SET state = 'expired', expired_at = v_token_expires_at,
            updated_at = v_now,
            revision = v_token_revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_token_id;
        UPDATE iam_invitations
        SET state = 'expired', expired_at = v_invitation_expires_at,
            updated_at = v_now,
            revision = v_invitation_revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_invitation_id;
        RETURN 'expired';
    END IF;
    UPDATE iam_invitations
    SET accepted_user_id = p_accepted_user_id,
        accepted_identifier_id = p_accepted_identifier_id,
        accepted_identifier_kind = 'email_v1',
        state = 'accepted', accepted_at = v_now, updated_at = v_now,
        revision = v_invitation_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_invitation_id;
    UPDATE iam_one_time_tokens
    SET state = 'consumed', consumed_at = v_now, updated_at = v_now,
        revision = v_token_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_token_id;
    RETURN 'accepted';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_accept_v1(
    name,uuid,bigint,uuid,bigint,uuid,uuid
) FROM PUBLIC;
