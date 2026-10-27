-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_register_accept_v1(
    p_scope_schema name,
    p_operation_id uuid, p_request_sha256 bytea,
    p_invitation_id uuid, p_expected_invitation_revision bigint,
    p_token_id uuid, p_expected_token_revision bigint, p_selector bytea,
    p_hash_key_id text, p_secret_hmac bytea, p_binding_sha256 bytea,
    p_user_id uuid, p_identifier_id uuid, p_access_mode_id uuid,
    p_role_assignment_id uuid, p_membership_id uuid,
    p_profile_schema_id uuid, p_profile_document bytea, p_profile_sha256 bytea,
    p_password_salt bytea, p_password_tag bytea,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_code text, result_revision bigint,
    authorization_clock bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_state text; v_revision bigint; v_expires_at timestamptz;
    v_display text; v_normalized text; v_normalized_sha bytea;
    v_role_id uuid; v_scope_kind text; v_scope_id uuid; v_organization_id uuid;
    v_token_state text; v_token_revision bigint; v_selector bytea;
    v_key_id text; v_hmac bytea; v_binding bytea; v_values jsonb;
    v_receipt_command text; v_receipt_hash bytea; v_receipt_code text;
    v_receipt_revision bigint; v_receipt_clock bigint; v_receipt_id uuid;
    v_clock bigint; v_assignment uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR
       p_invitation_id IS NULL OR p_token_id IS NULL OR p_user_id IS NULL OR
       p_identifier_id IS NULL OR p_access_mode_id IS NULL OR
       p_role_assignment_id IS NULL OR p_profile_schema_id IS NULL OR
       p_expected_invitation_revision < 1 OR p_expected_token_revision < 1 OR
       p_expected_authorization_version <> 1 OR
       pg_catalog.octet_length(p_request_sha256) <> 32 OR
       pg_catalog.octet_length(p_selector) <> 16 OR
       NOT iam_hmac_key_id_valid_v1(p_hash_key_id) OR
       pg_catalog.octet_length(p_secret_hmac) <> 32 OR
       pg_catalog.octet_length(p_binding_sha256) <> 32 OR
       pg_catalog.octet_length(p_profile_sha256) <> 32 OR
       pg_catalog.octet_length(p_password_salt) <> 16 OR
       pg_catalog.octet_length(p_password_tag) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM invitation registration input is invalid';
    END IF;
    SELECT invitation.state, invitation.revision, invitation.expires_at,
           invitation.display_value, invitation.normalized_value,
           invitation.normalized_sha256, invitation.role_id,
           invitation.role_scope_kind, invitation.role_scope_id,
           invitation.organization_id, token.state, token.revision,
           token.selector, token.hash_key_id, token.secret_hmac,
           token.binding_sha256
    INTO v_state, v_revision, v_expires_at, v_display, v_normalized,
         v_normalized_sha, v_role_id, v_scope_kind, v_scope_id,
         v_organization_id, v_token_state, v_token_revision, v_selector,
         v_key_id, v_hmac, v_binding
    FROM iam_invitations AS invitation
    JOIN iam_one_time_tokens AS token
      ON token.id = p_token_id AND token.invitation_id = invitation.id
     AND token.purpose = 'invitation_accept_v1'
    WHERE invitation.id = p_invitation_id
    FOR UPDATE OF invitation, token;
    IF NOT FOUND OR v_selector IS DISTINCT FROM p_selector OR
       v_key_id IS DISTINCT FROM p_hash_key_id OR
       v_hmac IS DISTINCT FROM p_secret_hmac OR
       v_binding IS DISTINCT FROM p_binding_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation credential is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_code,
           receipt.result_revision, receipt.clock_value, receipt.result_id
    INTO v_receipt_command, v_receipt_hash, v_receipt_code,
         v_receipt_revision, v_receipt_clock, v_receipt_id
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_receipt_command IS DISTINCT FROM 'iam.invitation.accept' OR
           v_receipt_hash IS DISTINCT FROM p_request_sha256 OR
           v_receipt_id IS DISTINCT FROM p_invitation_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM invitation operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_receipt_code, v_receipt_revision,
            v_receipt_clock, true;
        RETURN;
    END IF;
    IF v_state IS DISTINCT FROM 'pending' OR
       v_revision IS DISTINCT FROM p_expected_invitation_revision OR
       v_token_state IS DISTINCT FROM 'pending' OR
       v_token_revision IS DISTINCT FROM p_expected_token_revision OR
       v_now >= v_expires_at OR EXISTS (
           SELECT 1 FROM iam_user_identifiers AS identifier
           WHERE identifier.kind = 'email_v1'
             AND identifier.normalized_value = v_normalized
             AND identifier.state <> 'retired'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM invitation registration CAS is stale';
    END IF;
    v_values := iam_identity_profile_assert_v1(
        'user', p_profile_schema_id, p_profile_document, p_profile_sha256
    );
    INSERT INTO iam_users (
        id, identity_kind, status, profile_resource_kind,
        profile_schema_version_id, profile_values, profile_sha256,
        token_version, authorization_version, revision, created_at, updated_at
    ) VALUES (
        p_user_id, 'human', 'active', 'user', p_profile_schema_id,
        v_values, p_profile_sha256, 1, 1, 1, v_now, v_now
    );
    INSERT INTO iam_user_identifiers (
        id, user_id, kind, display_value, normalized_value, normalized_sha256,
        state, revision, verified_at, activated_at, created_at, updated_at
    ) VALUES (
        p_identifier_id, p_user_id, 'email_v1', v_display, v_normalized,
        v_normalized_sha, 'active', 1, v_now, v_now, v_now, v_now
    );
    INSERT INTO iam_user_access_modes (
        id, user_id, mode, state, revision,
        created_at, updated_at, activated_at
    ) VALUES (
        p_access_mode_id, p_user_id, 'operator_console', 'active', 1,
        v_now, v_now, v_now
    );
    INSERT INTO iam_password_credentials (
        user_id, credential_version, state, algorithm, salt, tag,
        memory_kib, iterations, parallelism, legacy_bcrypt_hash,
        upgrade_required, revision, created_at, updated_at
    ) VALUES (
        p_user_id, 1, 'active', 'argon2id_v1', p_password_salt, p_password_tag,
        65536, 3, 4, NULL, false, 1, v_now, v_now
    );
    v_assignment := iam_invitation_grant_role_v1(
        p_user_id, p_role_assignment_id, p_membership_id, v_role_id,
        v_scope_kind, v_scope_id, v_organization_id, v_now
    );
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1, updated_at = v_now
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF v_assignment IS DISTINCT FROM p_role_assignment_id OR v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM invitation registration effect is unavailable';
    END IF;
    UPDATE iam_invitations SET accepted_user_id = p_user_id,
        accepted_identifier_id = p_identifier_id,
        accepted_identifier_kind = 'email_v1', state = 'accepted',
        accepted_at = v_now, updated_at = v_now, revision = v_revision + 1
    WHERE id = p_invitation_id;
    UPDATE iam_one_time_tokens SET state = 'consumed', consumed_at = v_now,
        updated_at = v_now, revision = v_token_revision + 1
    WHERE id = p_token_id;
    DELETE FROM iam_invitation_delivery_artifacts
    WHERE invitation_id = p_invitation_id AND token_id = p_token_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM invitation delivery artifact is stale';
    END IF;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value, result_code,
        result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.invitation.accept', p_request_sha256,
        v_clock, 'iam.invitation.accepted', p_invitation_id,
        v_revision + 1, p_request_sha256, v_now
    );
    RETURN QUERY SELECT 'iam.invitation.accepted'::text,
        v_revision + 1, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_register_accept_v1(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,
    bytea,bytea,uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bytea,bigint
) FROM PUBLIC;
