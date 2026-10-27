-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_rotate_active_v2(
    p_scope_schema name,
    p_token_id uuid,
    p_family_id uuid,
    p_session_id uuid,
    p_successor_id uuid,
    p_successor_selector bytea,
    p_successor_hash_key_id text,
    p_successor_secret_hmac bytea
)
RETURNS TABLE (
    issued_at timestamptz,
    locator_selector bytea,
    locator_expected_revision bigint,
    locator_revision bigint,
    locator_source_sha256 bytea,
    locator_state text,
    locator_projected_at timestamptz,
    locator_retain_until timestamptz,
    successor_selector bytea,
    successor_revision bigint,
    successor_source_sha256 bytea,
    successor_state text,
    successor_issued_at timestamptz,
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
    v_token_revision bigint;
    v_sequence bigint;
    v_selector bytea;
    v_secret_hmac bytea;
    v_auth_time timestamptz;
    v_acr text;
    v_amr_password boolean;
    v_amr_webauthn boolean;
    v_amr_totp boolean;
    v_amr_recovery boolean;
    v_step_up_sha256 bytea;
    v_family_revision bigint;
    v_rotation_sequence bigint;
    v_absolute_expires_at timestamptz;
    v_session_revision bigint;
    v_retain_until timestamptz;
    v_idle_expires_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_successor_id IS NULL OR
       pg_catalog.octet_length(p_successor_selector) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_successor_secret_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM active refresh rotation input is invalid';
    END IF;
    SELECT token.revision, token.sequence, token.selector,
           token.secret_hmac, token.auth_time, token.acr,
           token.amr_password, token.amr_webauthn, token.amr_totp,
           token.amr_recovery, token.step_up_purpose_sha256,
           family.revision, family.rotation_sequence,
           family.absolute_expires_at, family.retain_until,
           session.revision
    INTO v_token_revision, v_sequence, v_selector, v_secret_hmac,
         v_auth_time, v_acr, v_amr_password, v_amr_webauthn,
         v_amr_totp, v_amr_recovery, v_step_up_sha256,
         v_family_revision, v_rotation_sequence, v_absolute_expires_at,
         v_retain_until, v_session_revision
    FROM iam_refresh_tokens AS token
    JOIN iam_refresh_families AS family
      ON family.id OPERATOR(pg_catalog.=) token.family_id
    JOIN iam_sessions AS session
      ON session.id OPERATOR(pg_catalog.=) family.session_id
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
      AND family.id OPERATOR(pg_catalog.=) p_family_id
      AND session.id OPERATOR(pg_catalog.=) p_session_id
      AND token.state OPERATOR(pg_catalog.=) 'active'
      AND family.state OPERATOR(pg_catalog.=) 'active'
      AND session.state OPERATOR(pg_catalog.=) 'active';
    IF NOT FOUND OR v_rotation_sequence IS DISTINCT FROM
           v_sequence OPERATOR(pg_catalog.-) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM active refresh rotation state is stale';
    END IF;
    v_idle_expires_at := LEAST(
        v_now OPERATOR(pg_catalog.+) INTERVAL '30 days',
        v_absolute_expires_at
    );
    UPDATE iam_refresh_tokens
    SET state = 'consumed', consumed_at = v_now, updated_at = v_now,
        revision = v_token_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_token_id;
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
        v_sequence OPERATOR(pg_catalog.+) 1, v_auth_time, v_acr,
        v_amr_password, v_amr_webauthn, v_amr_totp, v_amr_recovery,
        v_step_up_sha256, 'active', 1, v_now, v_now,
        v_idle_expires_at, NULL, NULL, NULL, v_retain_until
    );
    UPDATE iam_refresh_families
    SET rotation_sequence = v_rotation_sequence OPERATOR(pg_catalog.+) 1,
        updated_at = v_now,
        revision = v_family_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_family_id;
    UPDATE iam_sessions
    SET last_seen_at = v_now, updated_at = v_now,
        revision = v_session_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_session_id;
    issued_at := v_now;
    locator_selector := v_selector;
    locator_expected_revision := v_token_revision;
    locator_revision := v_token_revision OPERATOR(pg_catalog.+) 1;
    locator_source_sha256 := v_secret_hmac;
    locator_state := 'consumed';
    locator_projected_at := v_now;
    locator_retain_until := v_retain_until;
    successor_selector := p_successor_selector;
    successor_revision := 1;
    successor_source_sha256 := p_successor_secret_hmac;
    successor_state := 'active';
    successor_issued_at := v_now;
    successor_projected_at := v_now;
    successor_family_expires_at := v_absolute_expires_at;
    successor_retain_until := v_retain_until;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_rotate_active_v2(
    name,uuid,uuid,uuid,uuid,bytea,text,bytea
) FROM PUBLIC;
