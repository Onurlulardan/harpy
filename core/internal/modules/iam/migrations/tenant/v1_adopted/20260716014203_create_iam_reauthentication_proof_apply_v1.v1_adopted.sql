-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reauthentication_proof_apply_v1(
    p_scope_schema name, p_user_id uuid, p_method text,
    p_authenticator_id uuid, p_expected_authenticator_revision bigint,
    p_expected_subtype_revision bigint,
    p_expected_password_version bigint,
    p_expected_password_revision bigint,
    p_webauthn_sign_count bigint, p_webauthn_backup_eligible boolean,
    p_webauthn_backup_state boolean, p_totp_time_step bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_authenticator_state text;
    v_authenticator_revision bigint;
    v_subtype_revision bigint;
    v_sign_count bigint;
    v_backup_eligible boolean;
    v_last_time_step bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_user_id IS NULL OR
       p_method NOT IN ('password_v1', 'webauthn_v1', 'totp_v1') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM reauthentication proof scope is invalid';
    END IF;
    IF p_method OPERATOR(pg_catalog.=) 'password_v1' THEN
        IF p_authenticator_id IS NOT NULL OR
           p_expected_password_version OPERATOR(pg_catalog.<=) 0 OR
           p_expected_password_revision OPERATOR(pg_catalog.<=) 0 THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'IAM password proof evidence is invalid';
        END IF;
        UPDATE iam_password_credentials
        SET last_used_at = v_now, updated_at = v_now,
            revision = revision OPERATOR(pg_catalog.+) 1
        WHERE user_id OPERATOR(pg_catalog.=) p_user_id
          AND state OPERATOR(pg_catalog.=) 'active'
          AND algorithm OPERATOR(pg_catalog.=) 'argon2id_v1'
          AND credential_version OPERATOR(pg_catalog.=)
              p_expected_password_version
          AND revision OPERATOR(pg_catalog.=) p_expected_password_revision;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM password proof CAS is stale';
        END IF;
        RETURN p_expected_password_revision OPERATOR(pg_catalog.+) 1;
    END IF;
    SELECT authenticator.state, authenticator.revision
    INTO v_authenticator_state, v_authenticator_revision
    FROM iam_authenticators AS authenticator
    WHERE authenticator.id OPERATOR(pg_catalog.=) p_authenticator_id
      AND authenticator.user_id OPERATOR(pg_catalog.=) p_user_id
      AND authenticator.kind OPERATOR(pg_catalog.=) p_method
    FOR UPDATE;
    IF v_authenticator_state IS DISTINCT FROM 'active' OR
       v_authenticator_revision IS DISTINCT FROM
           p_expected_authenticator_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM authenticator proof CAS is stale';
    END IF;
    IF p_method OPERATOR(pg_catalog.=) 'webauthn_v1' THEN
        SELECT credential.revision, credential.sign_count,
               credential.backup_eligible
        INTO v_subtype_revision, v_sign_count, v_backup_eligible
        FROM iam_webauthn_credentials AS credential
        WHERE credential.authenticator_id OPERATOR(pg_catalog.=)
            p_authenticator_id
        FOR UPDATE;
        IF v_subtype_revision IS DISTINCT FROM p_expected_subtype_revision OR
           p_webauthn_sign_count OPERATOR(pg_catalog.<) 0 OR
           p_webauthn_backup_eligible IS DISTINCT FROM v_backup_eligible OR
           (p_webauthn_backup_state AND NOT p_webauthn_backup_eligible) OR
           (v_sign_count OPERATOR(pg_catalog.>) 0 AND
            p_webauthn_sign_count OPERATOR(pg_catalog.>) 0 AND
            p_webauthn_sign_count OPERATOR(pg_catalog.<=) v_sign_count AND
            NOT (v_backup_eligible AND p_webauthn_backup_state)) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM WebAuthn proof CAS is stale';
        END IF;
        UPDATE iam_webauthn_credentials
        SET sign_count = CASE
                WHEN p_webauthn_sign_count OPERATOR(pg_catalog.>) sign_count
                THEN p_webauthn_sign_count ELSE sign_count END,
            backup_state = p_webauthn_backup_state, updated_at = v_now,
            revision = revision OPERATOR(pg_catalog.+) 1
        WHERE authenticator_id OPERATOR(pg_catalog.=) p_authenticator_id;
    ELSE
        SELECT totp.revision, totp.last_accepted_time_step
        INTO v_subtype_revision, v_last_time_step
        FROM iam_totp_authenticators AS totp
        WHERE totp.authenticator_id OPERATOR(pg_catalog.=) p_authenticator_id
          AND totp.profile OPERATOR(pg_catalog.=) 'totp_sha256_6_30_v1'
          AND totp.verified_at IS NOT NULL
        FOR UPDATE;
        IF v_subtype_revision IS DISTINCT FROM p_expected_subtype_revision OR
           p_totp_time_step OPERATOR(pg_catalog.<) 0 OR
           (v_last_time_step IS NOT NULL AND
            p_totp_time_step OPERATOR(pg_catalog.<=) v_last_time_step) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM TOTP proof CAS is stale';
        END IF;
        UPDATE iam_totp_authenticators
        SET last_accepted_time_step = p_totp_time_step, updated_at = v_now,
            revision = revision OPERATOR(pg_catalog.+) 1
        WHERE authenticator_id OPERATOR(pg_catalog.=) p_authenticator_id;
    END IF;
    UPDATE iam_authenticators
    SET last_used_at = v_now, updated_at = v_now,
        revision = revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_authenticator_id
      AND revision OPERATOR(pg_catalog.=) p_expected_authenticator_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM authenticator proof update is stale';
    END IF;
    RETURN p_expected_authenticator_revision OPERATOR(pg_catalog.+) 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauthentication_proof_apply_v1(
    name,uuid,text,uuid,bigint,bigint,bigint,bigint,bigint,boolean,
    boolean,bigint
) FROM PUBLIC;
