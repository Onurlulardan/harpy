-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_initial_password_start_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_user_id uuid,
    p_installation_operation_id uuid,
    p_credential_id uuid,
    p_credential_version bigint,
    p_credential_revision bigint,
    p_selector bytea,
    p_secret_key_id text,
    p_secret_hmac bytea,
    p_csrf_key_id text,
    p_csrf_hmac bytea
)
RETURNS TABLE (
    revision bigint,
    issued_at timestamptz,
    expires_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_existing iam_initial_password_continuations%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM initial password start scope is invalid';
    END IF;
    IF p_tenant_id IS NULL OR p_user_id IS NULL OR
       p_installation_operation_id IS NULL OR p_credential_id IS NULL OR
       p_credential_version OPERATOR(pg_catalog.<) 1 OR
       p_credential_revision OPERATOR(pg_catalog.<) 1 OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       pg_catalog.char_length(p_secret_key_id) NOT BETWEEN 16 AND 64 OR
       pg_catalog.char_length(p_csrf_key_id) NOT BETWEEN 16 AND 64 OR
       p_secret_key_id IS NOT DISTINCT FROM p_csrf_key_id OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_csrf_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM initial password start evidence is invalid';
    END IF;
    SELECT continuation.*
    INTO v_existing
    FROM iam_initial_password_continuations AS continuation
    WHERE continuation.selector OPERATOR(pg_catalog.=) p_selector;
    IF FOUND THEN
        IF v_existing.tenant_id IS DISTINCT FROM p_tenant_id OR
           v_existing.user_id IS DISTINCT FROM p_user_id OR
           v_existing.installation_operation_id IS DISTINCT FROM
                p_installation_operation_id OR
           v_existing.credential_id IS DISTINCT FROM p_credential_id OR
           v_existing.credential_version IS DISTINCT FROM p_credential_version OR
           v_existing.credential_revision IS DISTINCT FROM p_credential_revision OR
           v_existing.secret_key_id IS DISTINCT FROM p_secret_key_id OR
           v_existing.secret_hmac IS DISTINCT FROM p_secret_hmac OR
           v_existing.csrf_key_id IS DISTINCT FROM p_csrf_key_id OR
           v_existing.csrf_hmac IS DISTINCT FROM p_csrf_hmac OR
           v_existing.state IS DISTINCT FROM 'active' THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM initial password start replay conflicts';
        END IF;
        RETURN QUERY SELECT v_existing.revision, v_existing.issued_at,
            v_existing.expires_at;
        RETURN;
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM iam_installation_password_credentials AS marker
        JOIN iam_password_credentials AS credential
          ON credential.user_id OPERATOR(pg_catalog.=) marker.user_id
        WHERE marker.credential_id OPERATOR(pg_catalog.=) p_credential_id AND
              marker.user_id OPERATOR(pg_catalog.=) p_user_id AND
              marker.installation_operation_id OPERATOR(pg_catalog.=)
                  p_installation_operation_id AND
              marker.credential_version OPERATOR(pg_catalog.=)
                  p_credential_version AND
              marker.state OPERATOR(pg_catalog.=) 'active' AND
              credential.credential_version OPERATOR(pg_catalog.=)
                  p_credential_version AND
              credential.revision OPERATOR(pg_catalog.=)
                  p_credential_revision AND
              credential.state OPERATOR(pg_catalog.=) 'reset_required'
        FOR UPDATE OF marker, credential
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM initial password credential changed';
    END IF;
    UPDATE iam_initial_password_continuations AS continuation
    SET state = CASE WHEN continuation.expires_at OPERATOR(pg_catalog.<=) v_now
            THEN 'expired' ELSE 'superseded' END,
        revision = continuation.revision OPERATOR(pg_catalog.+) 1,
        updated_at = v_now, consumed_at = v_now
    WHERE continuation.user_id OPERATOR(pg_catalog.=) p_user_id AND
          continuation.state OPERATOR(pg_catalog.=) 'active';
    INSERT INTO iam_initial_password_continuations (
        selector, tenant_id, user_id, installation_operation_id,
        credential_id, credential_version, credential_revision,
        secret_key_id, secret_hmac, csrf_key_id, csrf_hmac,
        state, policy_attempts, revision, issued_at, expires_at,
        updated_at, consumed_at
    ) VALUES (
        p_selector, p_tenant_id, p_user_id, p_installation_operation_id,
        p_credential_id, p_credential_version, p_credential_revision,
        p_secret_key_id, p_secret_hmac, p_csrf_key_id, p_csrf_hmac,
        'active', 0, 1, v_now,
        v_now OPERATOR(pg_catalog.+) INTERVAL '10 minutes', v_now, NULL
    );
    RETURN QUERY SELECT 1::bigint, v_now,
        v_now OPERATOR(pg_catalog.+) INTERVAL '10 minutes';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_password_start_v1(
    name,uuid,uuid,uuid,uuid,bigint,bigint,bytea,text,bytea,text,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_password_start_v1(
    name,uuid,uuid,uuid,uuid,bigint,bigint,bytea,text,bytea,text,bytea
) TO harpy_runtime;
