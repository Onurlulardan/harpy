-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_accept_v2(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_invitation_id uuid,
    p_expected_invitation_revision bigint,
    p_token_id uuid,
    p_expected_token_revision bigint,
    p_selector bytea,
    p_hash_key_id text,
    p_secret_hmac bytea,
    p_binding_sha256 bytea,
    p_accepted_user_id uuid,
    p_accepted_identifier_id uuid,
    p_role_assignment_id uuid,
    p_membership_id uuid,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_code text, result_revision bigint,
    authorization_clock bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_invitation_state text;
    v_invitation_revision bigint;
    v_invitation_expires_at timestamptz;
    v_normalized_value text;
    v_role_id uuid;
    v_role_scope_kind text;
    v_role_scope_id uuid;
    v_organization_id uuid;
    v_token_state text;
    v_token_revision bigint;
    v_token_selector bytea;
    v_token_key_id text;
    v_token_hmac bytea;
    v_token_binding bytea;
    v_user_state text;
    v_user_authorization_version bigint;
    v_identifier_state text;
    v_identifier_value text;
    v_receipt_command text;
    v_receipt_hash bytea;
    v_receipt_code text;
    v_receipt_revision bigint;
    v_receipt_clock bigint;
    v_receipt_result_id uuid;
    v_clock bigint;
    v_assignment_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation acceptance scope is invalid';
    END IF;
    IF p_operation_id IS NULL OR p_invitation_id IS NULL OR p_token_id IS NULL OR
       p_accepted_user_id IS NULL OR p_accepted_identifier_id IS NULL OR
       p_role_assignment_id IS NULL OR p_expected_invitation_revision < 1 OR
       p_expected_token_revision < 1 OR p_expected_authorization_version < 1 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       p_hash_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_binding_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM invitation acceptance input is invalid';
    END IF;
    SELECT invitation.state, invitation.revision, invitation.expires_at,
           invitation.normalized_value, invitation.role_id,
           invitation.role_scope_kind, invitation.role_scope_id,
           invitation.organization_id, token.state, token.revision,
           token.selector, token.hash_key_id, token.secret_hmac,
           token.binding_sha256, principal.status,
           principal.authorization_version, identifier.state,
           identifier.normalized_value
    INTO v_invitation_state, v_invitation_revision, v_invitation_expires_at,
         v_normalized_value, v_role_id, v_role_scope_kind,
         v_role_scope_id, v_organization_id, v_token_state,
         v_token_revision, v_token_selector, v_token_key_id,
         v_token_hmac, v_token_binding, v_user_state,
         v_user_authorization_version, v_identifier_state, v_identifier_value
    FROM iam_invitations AS invitation
    JOIN iam_one_time_tokens AS token
      ON token.id = p_token_id AND token.invitation_id = invitation.id
     AND token.purpose = 'invitation_accept_v1'
    JOIN iam_users AS principal ON principal.id = p_accepted_user_id
    JOIN iam_user_identifiers AS identifier
      ON identifier.id = p_accepted_identifier_id
     AND identifier.user_id = principal.id AND identifier.kind = 'email_v1'
    WHERE invitation.id = p_invitation_id
    FOR UPDATE OF invitation, token, principal, identifier;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM invitation acceptance target is unknown';
    END IF;
    IF v_token_selector IS DISTINCT FROM p_selector OR
       v_token_key_id IS DISTINCT FROM p_hash_key_id OR
       v_token_hmac IS DISTINCT FROM p_secret_hmac OR
       v_token_binding IS DISTINCT FROM p_binding_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation credential is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_code,
           receipt.result_revision, receipt.clock_value, receipt.result_id
    INTO v_receipt_command, v_receipt_hash, v_receipt_code,
         v_receipt_revision, v_receipt_clock, v_receipt_result_id
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_receipt_command IS DISTINCT FROM 'iam.invitation.accept' OR
           v_receipt_hash IS DISTINCT FROM p_request_sha256 OR
           v_receipt_result_id IS DISTINCT FROM p_invitation_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM invitation operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_receipt_code, v_receipt_revision,
            v_receipt_clock, true;
        RETURN;
    END IF;
    IF v_invitation_state IS DISTINCT FROM 'pending' OR
       v_invitation_revision IS DISTINCT FROM p_expected_invitation_revision OR
       v_token_state IS DISTINCT FROM 'pending' OR
       v_token_revision IS DISTINCT FROM p_expected_token_revision OR
       v_now >= v_invitation_expires_at OR v_user_state IS DISTINCT FROM 'active' OR
       v_user_authorization_version IS DISTINCT FROM
           p_expected_authorization_version OR
       v_identifier_state IS DISTINCT FROM 'active' OR
       v_identifier_value IS DISTINCT FROM v_normalized_value THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM invitation acceptance CAS is stale';
    END IF;
    v_assignment_id := iam_invitation_grant_role_v1(
        p_accepted_user_id, p_role_assignment_id, p_membership_id,
        v_role_id, v_role_scope_kind, v_role_scope_id,
        v_organization_id, v_now
    );
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1, updated_at = v_now
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL OR v_assignment_id IS DISTINCT FROM p_role_assignment_id THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM invitation authorization clock is unavailable';
    END IF;
    UPDATE iam_users
    SET authorization_version = authorization_version + 1,
        revision = revision + 1, updated_at = v_now
    WHERE id = p_accepted_user_id;
    UPDATE iam_invitations
    SET accepted_user_id = p_accepted_user_id,
        accepted_identifier_id = p_accepted_identifier_id,
        accepted_identifier_kind = 'email_v1', state = 'accepted',
        accepted_at = v_now, updated_at = v_now,
        revision = v_invitation_revision + 1
    WHERE id = p_invitation_id;
    UPDATE iam_one_time_tokens
    SET state = 'consumed', consumed_at = v_now,
        updated_at = v_now, revision = v_token_revision + 1
    WHERE id = p_token_id;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value, result_code,
        result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.invitation.accept', p_request_sha256,
        v_clock, 'iam.invitation.accepted', p_invitation_id,
        v_invitation_revision + 1, p_request_sha256, v_now
    );
    RETURN QUERY SELECT 'iam.invitation.accepted'::text,
        v_invitation_revision + 1, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_accept_v2(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,bytea,bytea,
    uuid,uuid,uuid,uuid,bigint
) FROM PUBLIC;
