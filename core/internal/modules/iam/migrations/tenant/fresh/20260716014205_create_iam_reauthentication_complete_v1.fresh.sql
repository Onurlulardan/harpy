-- Reauthentication completion boundary.
-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reauthentication_complete_v1(
    p_scope_schema name, p_tenant_id uuid, p_scope_binding bytea,
    p_challenge_id uuid, p_expected_challenge_revision bigint,
    p_expected_attempt_count smallint, p_user_id uuid, p_session_id uuid,
    p_expected_user_revision bigint, p_expected_session_revision bigint,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint, p_access_mode_id uuid,
    p_access_mode text, p_expected_access_mode_revision bigint,
    p_portal_binding_id uuid, p_expected_portal_binding_revision bigint,
    p_family_id uuid, p_token_id uuid, p_expected_family_revision bigint,
    p_expected_token_revision bigint, p_authenticator_id uuid,
    p_expected_authenticator_revision bigint,
    p_expected_subtype_revision bigint,
    p_expected_password_version bigint,
    p_expected_password_revision bigint,
    p_purpose_sha256 bytea, p_context_sha256 bytea, p_method text,
    p_webauthn_sign_count bigint, p_webauthn_backup_eligible boolean,
    p_webauthn_backup_state boolean, p_totp_time_step bigint,
    p_successor_id uuid, p_successor_selector bytea,
    p_successor_hash_key_id text, p_successor_secret_hmac bytea
)
RETURNS TABLE (
    issued_at timestamptz, challenge_revision bigint,
    family_revision bigint, session_revision bigint,
    consumed_token_revision bigint, successor_token_revision bigint,
    locator_selector bytea, locator_expected_revision bigint,
    locator_revision bigint, locator_source_sha256 bytea,
    locator_state text, locator_projected_at timestamptz,
    locator_retain_until timestamptz, successor_selector bytea,
    successor_revision bigint, successor_source_sha256 bytea,
    successor_state text, successor_issued_at timestamptz,
    successor_projected_at timestamptz,
    successor_family_expires_at timestamptz,
    successor_retain_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_challenge_state text;
    v_challenge_revision bigint;
    v_attempt_count smallint;
    v_challenge_user_id uuid;
    v_challenge_session_id uuid;
    v_challenge_authenticator_id uuid;
    v_challenge_method text;
    v_challenge_purpose bytea;
    v_challenge_context bytea;
    v_challenge_expires_at timestamptz;
    v_resolved text;
    v_proof_revision bigint;
    v_issued_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_scope_binding) IS DISTINCT FROM 32 OR
       iam_reauth_binding_sha256_v1(
           p_tenant_id, p_user_id, p_session_id, p_expected_token_version,
           p_expected_authorization_version, p_access_mode,
           p_portal_binding_id
       ) IS DISTINCT FROM p_scope_binding OR
       pg_catalog.octet_length(p_purpose_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_context_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_successor_selector) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_successor_secret_hmac) IS DISTINCT FROM 32
       THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM reauthentication completion scope is invalid';
    END IF;
    SELECT challenge.state, challenge.revision, challenge.attempt_count,
           challenge.user_id, challenge.session_id,
           challenge.authenticator_id, challenge.proof_kind,
           challenge.step_up_purpose_sha256, challenge.context_sha256,
           challenge.expires_at
    INTO v_challenge_state, v_challenge_revision, v_attempt_count,
         v_challenge_user_id, v_challenge_session_id,
         v_challenge_authenticator_id, v_challenge_method,
         v_challenge_purpose, v_challenge_context, v_challenge_expires_at
    FROM iam_auth_challenges AS challenge
    WHERE challenge.id OPERATOR(pg_catalog.=) p_challenge_id
      AND challenge.purpose OPERATOR(pg_catalog.=) 'reauth_v1'
    FOR UPDATE;
    IF v_challenge_state IS DISTINCT FROM 'pending' OR
       v_challenge_revision IS DISTINCT FROM p_expected_challenge_revision OR
       v_attempt_count IS DISTINCT FROM p_expected_attempt_count OR
       v_challenge_user_id IS DISTINCT FROM p_user_id OR
       v_challenge_session_id IS DISTINCT FROM p_session_id OR
       v_challenge_authenticator_id IS DISTINCT FROM p_authenticator_id OR
       v_challenge_method IS DISTINCT FROM p_method OR
       v_challenge_purpose IS DISTINCT FROM p_purpose_sha256 OR
       v_challenge_context IS DISTINCT FROM p_context_sha256 OR
       v_challenge_expires_at OPERATOR(pg_catalog.<=) v_now THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication completion CAS is stale';
    END IF;
    v_resolved := iam_challenge_resolve_v1(
        p_scope_schema, p_challenge_id, p_expected_challenge_revision,
        p_expected_attempt_count, 'consume'
    );
    IF v_resolved IS DISTINCT FROM 'consumed' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication challenge was not consumed';
    END IF;
    v_proof_revision := iam_reauthentication_proof_apply_v1(
        p_scope_schema, p_user_id, p_method, p_authenticator_id,
        p_expected_authenticator_revision, p_expected_subtype_revision,
        p_expected_password_version, p_expected_password_revision,
        p_webauthn_sign_count, p_webauthn_backup_eligible,
        p_webauthn_backup_state, p_totp_time_step
    );
    SELECT rotated.issued_at, rotated.family_revision,
           rotated.session_revision, rotated.locator_selector,
           rotated.locator_expected_revision, rotated.locator_revision,
           rotated.locator_source_sha256, rotated.locator_state,
           rotated.locator_projected_at, rotated.locator_retain_until,
           rotated.successor_selector, rotated.successor_revision,
           rotated.successor_source_sha256, rotated.successor_state,
           rotated.successor_issued_at, rotated.successor_projected_at,
           rotated.successor_family_expires_at,
           rotated.successor_retain_until
    INTO v_issued_at, family_revision, session_revision, locator_selector,
         locator_expected_revision, locator_revision, locator_source_sha256,
         locator_state, locator_projected_at, locator_retain_until,
         successor_selector, successor_revision, successor_source_sha256,
         successor_state, successor_issued_at, successor_projected_at,
         successor_family_expires_at, successor_retain_until
    FROM iam_reauthentication_rotate_active_v1(
        p_scope_schema, p_token_id, p_family_id,
        p_expected_token_revision, p_expected_family_revision,
        p_session_id, p_expected_session_revision, p_user_id,
        p_expected_user_revision, p_expected_token_version,
        p_expected_authorization_version, p_access_mode_id,
        p_expected_access_mode_revision, p_portal_binding_id,
        p_expected_portal_binding_revision, p_successor_id,
        p_successor_selector, p_successor_hash_key_id,
        p_successor_secret_hmac, p_method, p_purpose_sha256
    ) AS rotated;
    challenge_revision := p_expected_challenge_revision
        OPERATOR(pg_catalog.+) 1;
    consumed_token_revision := locator_revision;
    successor_token_revision := successor_revision;
    issued_at := v_issued_at;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauthentication_complete_v1(
    name,uuid,bytea,uuid,bigint,smallint,uuid,uuid,bigint,bigint,bigint,
    bigint,uuid,text,bigint,uuid,bigint,uuid,uuid,bigint,bigint,uuid,
    bigint,bigint,bigint,bigint,bytea,bytea,text,bigint,boolean,boolean,
    bigint,uuid,bytea,text,bytea
) FROM PUBLIC;
