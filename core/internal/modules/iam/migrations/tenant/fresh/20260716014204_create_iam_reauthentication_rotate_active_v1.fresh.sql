-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reauthentication_rotate_active_v1(
    p_scope_schema name, p_token_id uuid, p_family_id uuid,
    p_expected_token_revision bigint, p_expected_family_revision bigint,
    p_session_id uuid, p_expected_session_revision bigint,
    p_user_id uuid, p_expected_user_revision bigint,
    p_expected_token_version bigint, p_expected_authorization_version bigint,
    p_access_mode_id uuid, p_expected_access_mode_revision bigint,
    p_portal_binding_id uuid, p_expected_portal_binding_revision bigint,
    p_successor_id uuid, p_successor_selector bytea,
    p_successor_hash_key_id text, p_successor_secret_hmac bytea,
    p_method text, p_purpose_sha256 bytea
)
RETURNS TABLE (
    issued_at timestamptz, family_revision bigint, session_revision bigint,
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
    v_decision text;
    v_sequence bigint;
    v_selector bytea;
    v_secret_hmac bytea;
    v_rotation_sequence bigint;
    v_absolute_expires_at timestamptz;
    v_retain_until timestamptz;
    v_idle_expires_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_successor_id IS NULL OR
       p_method NOT IN ('password_v1', 'webauthn_v1', 'totp_v1') OR
       pg_catalog.octet_length(p_purpose_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_successor_selector) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_successor_secret_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM reauthentication rotation input is invalid';
    END IF;
    v_decision := iam_refresh_decide_v2(
        p_scope_schema, p_token_id, p_family_id,
        p_expected_token_revision, p_expected_family_revision,
        p_session_id, p_expected_session_revision, p_user_id,
        p_expected_user_revision, p_expected_token_version,
        p_expected_authorization_version, p_access_mode_id,
        p_expected_access_mode_revision, p_portal_binding_id,
        p_expected_portal_binding_revision
    );
    IF v_decision IS DISTINCT FROM 'rotate' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication rotation state is stale';
    END IF;
    SELECT token.sequence, token.selector, token.secret_hmac,
           family.rotation_sequence, family.absolute_expires_at,
           family.retain_until
    INTO STRICT v_sequence, v_selector, v_secret_hmac,
         v_rotation_sequence, v_absolute_expires_at, v_retain_until
    FROM iam_refresh_tokens AS token
    JOIN iam_refresh_families AS family
      ON family.id OPERATOR(pg_catalog.=) token.family_id
    JOIN iam_sessions AS session
      ON session.id OPERATOR(pg_catalog.=) family.session_id
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
      AND family.id OPERATOR(pg_catalog.=) p_family_id
      AND session.id OPERATOR(pg_catalog.=) p_session_id;
    IF v_rotation_sequence IS DISTINCT FROM
       v_sequence OPERATOR(pg_catalog.-) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication rotation sequence is stale';
    END IF;
    v_idle_expires_at := LEAST(
        v_now OPERATOR(pg_catalog.+) INTERVAL '30 days',
        v_absolute_expires_at
    );
    UPDATE iam_refresh_tokens
    SET state = 'consumed', consumed_at = v_now, updated_at = v_now,
        revision = p_expected_token_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_token_id
      AND family_id OPERATOR(pg_catalog.=) p_family_id
      AND state OPERATOR(pg_catalog.=) 'active'
      AND revision OPERATOR(pg_catalog.=) p_expected_token_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication token CAS is stale';
    END IF;
    INSERT INTO iam_refresh_tokens (
        id, family_id, family_absolute_expires_at, parent_id, parent_sequence,
        selector, hash_key_id, secret_hmac, sequence, auth_time, acr,
        amr_password, amr_webauthn, amr_totp, amr_recovery,
        step_up_purpose_sha256, state, revision, issued_at, updated_at,
        idle_expires_at, consumed_at, revoked_at, expired_at, retain_until
    ) VALUES (
        p_successor_id, p_family_id, v_absolute_expires_at,
        p_token_id, v_sequence, p_successor_selector,
        p_successor_hash_key_id, p_successor_secret_hmac,
        v_sequence OPERATOR(pg_catalog.+) 1, v_now, 'step_up',
        p_method OPERATOR(pg_catalog.=) 'password_v1',
        p_method OPERATOR(pg_catalog.=) 'webauthn_v1',
        p_method OPERATOR(pg_catalog.=) 'totp_v1', false,
        p_purpose_sha256, 'active', 1, v_now, v_now,
        v_idle_expires_at, NULL, NULL, NULL, v_retain_until
    );
    UPDATE iam_refresh_families
    SET rotation_sequence = v_rotation_sequence OPERATOR(pg_catalog.+) 1,
        updated_at = v_now,
        revision = p_expected_family_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_family_id
      AND state OPERATOR(pg_catalog.=) 'active'
      AND revision OPERATOR(pg_catalog.=) p_expected_family_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication family CAS is stale';
    END IF;
    UPDATE iam_sessions
    SET last_seen_at = v_now, updated_at = v_now,
        revision = p_expected_session_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_session_id
      AND state OPERATOR(pg_catalog.=) 'active'
      AND revision OPERATOR(pg_catalog.=) p_expected_session_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication session CAS is stale';
    END IF;
    issued_at := v_now;
    family_revision := p_expected_family_revision OPERATOR(pg_catalog.+) 1;
    session_revision := p_expected_session_revision OPERATOR(pg_catalog.+) 1;
    locator_selector := v_selector;
    locator_expected_revision := p_expected_token_revision;
    locator_revision := p_expected_token_revision OPERATOR(pg_catalog.+) 1;
    locator_state := 'consumed';
    locator_projected_at := v_now;
    locator_retain_until := v_retain_until;
    locator_source_sha256 := iam_refresh_locator_source_sha256_v1(
        p_token_id, p_family_id, v_selector, locator_revision,
        locator_state, v_secret_hmac
    );
    successor_selector := p_successor_selector;
    successor_revision := 1;
    successor_state := 'active';
    successor_issued_at := v_now;
    successor_projected_at := v_now;
    successor_family_expires_at := v_absolute_expires_at;
    successor_retain_until := v_retain_until;
    successor_source_sha256 := iam_refresh_locator_source_sha256_v1(
        p_successor_id, p_family_id, p_successor_selector,
        successor_revision, successor_state, p_successor_secret_hmac
    );
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauthentication_rotate_active_v1(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea,text,bytea
) FROM PUBLIC;
