-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_rotate_v1(
    p_scope_schema name,
    p_token_id uuid,
    p_expected_revision bigint,
    p_successor_id uuid,
    p_successor_selector bytea,
    p_successor_hash_key_id text,
    p_successor_secret_hmac bytea
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_token_state text;
    v_token_revision bigint;
    v_sequence bigint;
    v_parent_issued_at timestamptz;
    v_family_id uuid;
    v_family_state text;
    v_family_revision bigint;
    v_family_expires_at timestamptz;
    v_session_id uuid;
    v_session_state text;
    v_session_revision bigint;
    v_user_id uuid;
    v_user_revision bigint;
    v_idle_expires_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh rotation scope is invalid';
    END IF;
    SELECT token.state, token.revision, token.sequence, token.issued_at,
           family.id, family.state, family.revision,
           family.absolute_expires_at, family.session_id,
           session.state, session.revision, family.user_id, principal.revision
    INTO v_token_state, v_token_revision, v_sequence, v_parent_issued_at,
         v_family_id, v_family_state, v_family_revision,
         v_family_expires_at, v_session_id,
         v_session_state, v_session_revision, v_user_id, v_user_revision
    FROM iam_refresh_tokens AS token
    JOIN iam_refresh_families AS family
      ON family.id OPERATOR(pg_catalog.=) token.family_id
    JOIN iam_sessions AS session
      ON session.id OPERATOR(pg_catalog.=) family.session_id
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) family.user_id
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
    FOR UPDATE OF principal, session, family, token;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM refresh token is unknown';
    END IF;
    IF v_token_state IS DISTINCT FROM 'active' THEN
        IF v_token_state IS NOT DISTINCT FROM 'consumed' AND
           v_family_state IS NOT DISTINCT FROM 'active' THEN
            UPDATE iam_refresh_tokens
            SET state = 'revoked',
                revoked_at = CASE
                    WHEN updated_at OPERATOR(pg_catalog.>) v_now
                        THEN updated_at ELSE v_now
                END,
                updated_at = CASE
                    WHEN updated_at OPERATOR(pg_catalog.>) v_now
                        THEN updated_at ELSE v_now
                END,
                revision = revision OPERATOR(pg_catalog.+) 1
            WHERE family_id OPERATOR(pg_catalog.=) v_family_id
              AND state OPERATOR(pg_catalog.=) 'active';
            UPDATE iam_refresh_families
            SET state = 'compromised', revoke_reason = 'reuse_detected',
                compromised_at = CASE
                    WHEN updated_at OPERATOR(pg_catalog.>) v_now
                        THEN updated_at ELSE v_now
                END,
                updated_at = CASE
                    WHEN updated_at OPERATOR(pg_catalog.>) v_now
                        THEN updated_at ELSE v_now
                END,
                revision = v_family_revision OPERATOR(pg_catalog.+) 1
            WHERE id OPERATOR(pg_catalog.=) v_family_id;
            IF v_session_state IS NOT DISTINCT FROM 'active' THEN
                UPDATE iam_sessions
                SET state = 'revoked',
                    revoked_at = CASE
                        WHEN updated_at OPERATOR(pg_catalog.>) v_now
                            THEN updated_at ELSE v_now
                    END,
                    updated_at = CASE
                        WHEN updated_at OPERATOR(pg_catalog.>) v_now
                            THEN updated_at ELSE v_now
                    END,
                    revision = v_session_revision OPERATOR(pg_catalog.+) 1
                WHERE id OPERATOR(pg_catalog.=) v_session_id;
            END IF;
            UPDATE iam_users
            SET token_version = token_version OPERATOR(pg_catalog.+) 1,
                updated_at = CASE
                    WHEN updated_at OPERATOR(pg_catalog.>) v_now
                        THEN updated_at ELSE v_now
                END,
                revision = v_user_revision OPERATOR(pg_catalog.+) 1
            WHERE id OPERATOR(pg_catalog.=) v_user_id;
        END IF;
        IF v_token_state IS NOT DISTINCT FROM 'expired' THEN
            RETURN 'expired';
        END IF;
        RETURN 'reused';
    END IF;
    IF v_token_revision IS DISTINCT FROM p_expected_revision OR
       v_family_state IS DISTINCT FROM 'active' OR
       v_session_state IS DISTINCT FROM 'active' OR
       v_now OPERATOR(pg_catalog.<=) v_parent_issued_at OR
       v_now OPERATOR(pg_catalog.>=) v_family_expires_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM refresh rotation CAS is stale';
    END IF;
    v_idle_expires_at := v_now OPERATOR(pg_catalog.+) INTERVAL '30 days';
    IF v_idle_expires_at OPERATOR(pg_catalog.>) v_family_expires_at THEN
        v_idle_expires_at := v_family_expires_at;
    END IF;
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
    )
    SELECT p_successor_id, token.family_id, token.family_absolute_expires_at,
           token.id, token.sequence, p_successor_selector,
           p_successor_hash_key_id, p_successor_secret_hmac,
           token.sequence OPERATOR(pg_catalog.+) 1,
           token.auth_time, token.acr, token.amr_password,
           token.amr_webauthn, token.amr_totp, token.amr_recovery,
           token.step_up_purpose_sha256, 'active', 1, v_now, v_now,
           v_idle_expires_at, NULL, NULL, NULL, token.retain_until
    FROM iam_refresh_tokens AS token
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id;
    UPDATE iam_refresh_families
    SET rotation_sequence = v_sequence OPERATOR(pg_catalog.+) 1,
        updated_at = v_now,
        revision = v_family_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) v_family_id;
    UPDATE iam_sessions
    SET last_seen_at = v_now, updated_at = v_now,
        revision = v_session_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) v_session_id;
    RETURN 'rotated';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_rotate_v1(
    name,uuid,bigint,uuid,bytea,text,bytea
) FROM PUBLIC;
