-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_session_issue_v1(
    p_scope_schema name,
    p_user_id uuid,
    p_expected_user_revision bigint,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_access_mode_id uuid,
    p_access_mode text,
    p_expected_access_mode_revision bigint,
    p_portal_binding_id uuid,
    p_expected_portal_binding_revision bigint,
    p_session_id uuid,
    p_device_kind text,
    p_browser_family text,
    p_os_family text,
    p_ip_hmac_key_id text,
    p_ip_hmac bytea,
    p_family_id uuid,
    p_token_id uuid,
    p_selector bytea,
    p_hash_key_id text,
    p_secret_hmac bytea
)
RETURNS TABLE (
    session_revision bigint,
    family_revision bigint,
    token_revision bigint,
    issued_at timestamptz,
    idle_expires_at timestamptz,
    absolute_expires_at timestamptz,
    retain_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_absolute_expires_at timestamptz;
    v_idle_expires_at timestamptz;
    v_token_retain_until timestamptz;
    v_user_status text;
    v_identity_kind text;
    v_user_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_user_updated_at timestamptz;
    v_access_state text;
    v_access_revision bigint;
    v_access_updated_at timestamptz;
    v_binding_id uuid;
    v_binding_state text;
    v_binding_revision bigint;
    v_binding_updated_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM session issue scope is invalid';
    END IF;
    IF p_user_id IS NOT DISTINCT FROM
           '00000000-0000-0000-0000-000000000000'::uuid OR
       p_access_mode_id IS NOT DISTINCT FROM
           '00000000-0000-0000-0000-000000000000'::uuid OR
       p_session_id IS NOT DISTINCT FROM
           '00000000-0000-0000-0000-000000000000'::uuid OR
       p_family_id IS NOT DISTINCT FROM
           '00000000-0000-0000-0000-000000000000'::uuid OR
       p_token_id IS NOT DISTINCT FROM
           '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_ip_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       NOT (
           (p_access_mode IS NOT DISTINCT FROM 'operator_console' AND
            p_portal_binding_id IS NULL AND
            p_expected_portal_binding_revision IS NULL) OR
           (p_access_mode IS NOT DISTINCT FROM 'portal' AND
            p_portal_binding_id IS NOT NULL AND
            p_expected_portal_binding_revision OPERATOR(pg_catalog.>) 0)
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM session issue input is invalid';
    END IF;
    SELECT principal.status, principal.identity_kind, principal.revision,
           principal.token_version, principal.authorization_version,
           principal.updated_at
    INTO v_user_status, v_identity_kind, v_user_revision,
         v_token_version, v_authorization_version, v_user_updated_at
    FROM iam_users AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_user_id
    FOR UPDATE;
    SELECT access.state, access.revision, access.updated_at
    INTO v_access_state, v_access_revision, v_access_updated_at
    FROM iam_user_access_modes AS access
    WHERE access.id OPERATOR(pg_catalog.=) p_access_mode_id
      AND access.user_id OPERATOR(pg_catalog.=) p_user_id
      AND access.mode OPERATOR(pg_catalog.=) p_access_mode
    FOR UPDATE;
    IF p_access_mode IS NOT DISTINCT FROM 'portal' THEN
        SELECT binding.id, binding.state, binding.revision, binding.updated_at
        INTO v_binding_id, v_binding_state, v_binding_revision,
             v_binding_updated_at
        FROM iam_portal_subject_bindings AS binding
        WHERE binding.id OPERATOR(pg_catalog.=) p_portal_binding_id
          AND binding.user_id OPERATOR(pg_catalog.=) p_user_id
          AND binding.access_mode_id OPERATOR(pg_catalog.=) p_access_mode_id
        FOR UPDATE;
    END IF;
    IF v_user_status IS NULL OR v_access_state IS NULL OR
       v_user_status IS DISTINCT FROM 'active' OR
       v_identity_kind IS DISTINCT FROM 'human' OR
       v_access_state IS DISTINCT FROM 'active' OR
       v_user_revision IS DISTINCT FROM p_expected_user_revision OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_authorization_version IS DISTINCT FROM
           p_expected_authorization_version OR
       v_access_revision IS DISTINCT FROM p_expected_access_mode_revision OR
       (p_access_mode IS NOT DISTINCT FROM 'portal' AND (
           v_binding_id IS NULL OR v_binding_state IS DISTINCT FROM 'active' OR
           v_binding_revision IS DISTINCT FROM
               p_expected_portal_binding_revision
       )) OR
       v_now OPERATOR(pg_catalog.<) v_user_updated_at OR
       v_now OPERATOR(pg_catalog.<) v_access_updated_at OR
       (v_binding_id IS NOT NULL AND
        v_now OPERATOR(pg_catalog.<) v_binding_updated_at) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM session issue CAS is stale';
    END IF;
    v_absolute_expires_at := v_now OPERATOR(pg_catalog.+) INTERVAL '90 days';
    v_idle_expires_at := v_now OPERATOR(pg_catalog.+) INTERVAL '30 days';
    v_token_retain_until :=
        v_absolute_expires_at OPERATOR(pg_catalog.+) INTERVAL '24 hours';
    INSERT INTO iam_sessions (
        id, user_id, access_mode_id, access_mode, portal_binding_id, state,
        device_kind, browser_family, os_family, ip_hmac_key_id, ip_hmac,
        revision, created_at, updated_at, last_seen_at, absolute_expires_at,
        revoked_at, expired_at, retain_until
    ) VALUES (
        p_session_id, p_user_id, p_access_mode_id, p_access_mode,
        p_portal_binding_id, 'active', p_device_kind, p_browser_family,
        p_os_family, p_ip_hmac_key_id, p_ip_hmac, 1, v_now, v_now, v_now,
        v_absolute_expires_at, NULL, NULL,
        v_absolute_expires_at OPERATOR(pg_catalog.+) INTERVAL '30 days'
    );
    INSERT INTO iam_refresh_families (
        id, user_id, session_id, state, rotation_sequence, revoke_reason,
        revision, created_at, updated_at, absolute_expires_at,
        revoked_at, compromised_at, expired_at, retain_until
    ) VALUES (
        p_family_id, p_user_id, p_session_id, 'active', 0, NULL,
        1, v_now, v_now, v_absolute_expires_at,
        NULL, NULL, NULL, v_token_retain_until
    );
    INSERT INTO iam_refresh_tokens (
        id, family_id, family_absolute_expires_at, parent_id, parent_sequence,
        selector, hash_key_id, secret_hmac, sequence, auth_time, acr,
        amr_password, amr_webauthn, amr_totp, amr_recovery,
        step_up_purpose_sha256, state, revision, issued_at, updated_at,
        idle_expires_at, consumed_at, revoked_at, expired_at, retain_until
    ) VALUES (
        p_token_id, p_family_id, v_absolute_expires_at, NULL, NULL,
        p_selector, p_hash_key_id, p_secret_hmac, 1, v_now, 'login',
        true, false, false, false, NULL, 'active', 1, v_now, v_now,
        v_idle_expires_at, NULL, NULL, NULL, v_token_retain_until
    );
    RETURN QUERY SELECT 1::bigint, 1::bigint, 1::bigint, v_now,
        v_idle_expires_at, v_absolute_expires_at, v_token_retain_until;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_session_issue_v1(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) FROM PUBLIC;
