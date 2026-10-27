-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_initial_password_complete_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_selector bytea,
    p_secret_hmac bytea,
    p_csrf_hmac bytea,
    p_expected_continuation_revision bigint,
    p_user_id uuid,
    p_credential_id uuid,
    p_expected_credential_version bigint,
    p_expected_credential_revision bigint,
    p_salt bytea,
    p_tag bytea,
    p_request_sha256 bytea
)
RETURNS TABLE (
    user_id uuid,
    access_mode_id uuid,
    user_revision bigint,
    token_version bigint,
    authorization_version bigint,
    access_mode_revision bigint,
    credential_version bigint,
    credential_revision bigint,
    continuation_revision bigint,
    installation_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_continuation iam_initial_password_continuations%ROWTYPE;
    v_marker iam_installation_password_credentials%ROWTYPE;
    v_credential_revision bigint;
    v_access_mode_id uuid;
    v_access_mode_revision bigint;
    v_user_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM initial password completion scope is invalid';
    END IF;
    IF p_tenant_id IS NULL OR p_user_id IS NULL OR p_credential_id IS NULL OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_csrf_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_salt) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_tag) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_expected_continuation_revision OPERATOR(pg_catalog.<) 1 OR
       p_expected_credential_version OPERATOR(pg_catalog.<) 1 OR
       p_expected_credential_revision OPERATOR(pg_catalog.<) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM initial password completion evidence is invalid';
    END IF;
    SELECT continuation.*
    INTO v_continuation
    FROM iam_initial_password_continuations AS continuation
    WHERE continuation.selector OPERATOR(pg_catalog.=) p_selector
    FOR UPDATE;
    IF NOT FOUND OR v_continuation.tenant_id IS DISTINCT FROM p_tenant_id OR
       v_continuation.user_id IS DISTINCT FROM p_user_id OR
       v_continuation.credential_id IS DISTINCT FROM p_credential_id OR
       v_continuation.credential_version IS DISTINCT FROM
            p_expected_credential_version OR
       v_continuation.credential_revision IS DISTINCT FROM
            p_expected_credential_revision OR
       v_continuation.state IS DISTINCT FROM 'active' OR
       v_continuation.revision IS DISTINCT FROM
            p_expected_continuation_revision OR
       v_continuation.secret_hmac IS DISTINCT FROM p_secret_hmac OR
       v_continuation.csrf_hmac IS DISTINCT FROM p_csrf_hmac OR
       v_continuation.policy_attempts OPERATOR(pg_catalog.>=) 5 OR
       v_now OPERATOR(pg_catalog.<) v_continuation.issued_at OR
       v_now OPERATOR(pg_catalog.>=) v_continuation.expires_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM initial password continuation changed';
    END IF;
    SELECT marker.*
    INTO v_marker
    FROM iam_installation_password_credentials AS marker
    WHERE marker.credential_id OPERATOR(pg_catalog.=) p_credential_id
    FOR UPDATE;
    IF NOT FOUND OR v_marker.user_id IS DISTINCT FROM p_user_id OR
       v_marker.installation_operation_id IS DISTINCT FROM
            v_continuation.installation_operation_id OR
       v_marker.credential_version IS DISTINCT FROM
            p_expected_credential_version OR
       v_marker.state IS DISTINCT FROM 'active' OR EXISTS (
            SELECT 1 FROM iam_sessions AS session
            WHERE session.user_id OPERATOR(pg_catalog.=) p_user_id AND
                  session.state OPERATOR(pg_catalog.=) 'active'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM installation credential changed';
    END IF;
    UPDATE iam_password_credentials AS credential
    SET credential_version =
            p_expected_credential_version OPERATOR(pg_catalog.+) 1,
        state = 'active', algorithm = 'argon2id_v1',
        salt = p_salt, tag = p_tag, memory_kib = 65536,
        iterations = 3, parallelism = 4, legacy_bcrypt_hash = NULL,
        upgrade_required = false,
        revision = credential.revision OPERATOR(pg_catalog.+) 1,
        updated_at = v_now, last_used_at = v_now, invalidated_at = NULL
    WHERE credential.user_id OPERATOR(pg_catalog.=) p_user_id AND
          credential.credential_version OPERATOR(pg_catalog.=)
              p_expected_credential_version AND
          credential.revision OPERATOR(pg_catalog.=)
              p_expected_credential_revision AND
          credential.state OPERATOR(pg_catalog.=) 'reset_required'
    RETURNING credential.revision INTO v_credential_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM initial password credential CAS changed';
    END IF;
    UPDATE iam_installation_password_credentials AS marker
    SET state = 'retired',
        revision = marker.revision OPERATOR(pg_catalog.+) 1,
        updated_at = v_now, retired_at = v_now
    WHERE marker.credential_id OPERATOR(pg_catalog.=) p_credential_id;
    UPDATE iam_initial_password_continuations AS continuation
    SET state = 'consumed',
        revision = continuation.revision OPERATOR(pg_catalog.+) 1,
        updated_at = v_now, consumed_at = v_now
    WHERE continuation.selector OPERATOR(pg_catalog.=) p_selector;
    SELECT recorded.token_version, recorded.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        v_continuation.installation_operation_id,
        'iam.initial_password.complete', p_request_sha256,
        p_user_id, p_user_id, v_credential_revision,
        'credential_target', p_user_id, 'password_credential', true, false
    ) AS recorded;
    SELECT principal.revision
    INTO STRICT v_user_revision
    FROM iam_users AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_user_id;
    SELECT access.id, access.revision
    INTO STRICT v_access_mode_id, v_access_mode_revision
    FROM iam_user_access_modes AS access
    WHERE access.user_id OPERATOR(pg_catalog.=) p_user_id AND
          access.mode OPERATOR(pg_catalog.=) 'operator_console' AND
          access.state OPERATOR(pg_catalog.=) 'active';
    RETURN QUERY SELECT p_user_id, v_access_mode_id, v_user_revision,
        v_token_version, v_authorization_version, v_access_mode_revision,
        p_expected_credential_version OPERATOR(pg_catalog.+) 1,
        v_credential_revision,
        p_expected_continuation_revision OPERATOR(pg_catalog.+) 1,
        v_marker.revision OPERATOR(pg_catalog.+) 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_password_complete_v1(
    name,uuid,bytea,bytea,bytea,bigint,uuid,uuid,bigint,bigint,
    bytea,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_password_complete_v1(
    name,uuid,bytea,bytea,bytea,bigint,uuid,uuid,bigint,bigint,
    bytea,bytea,bytea
) TO forgecrud_runtime;
