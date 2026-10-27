-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_initial_password_reject_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_selector bytea,
    p_secret_hmac bytea,
    p_csrf_hmac bytea,
    p_expected_revision bigint
)
RETURNS TABLE (
    state text,
    policy_attempts smallint,
    revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_continuation iam_initial_password_continuations%ROWTYPE;
    v_attempts smallint;
    v_state text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM initial password rejection scope is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_csrf_hmac) IS DISTINCT FROM 32 OR
       p_expected_revision OPERATOR(pg_catalog.<) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM initial password rejection evidence is invalid';
    END IF;
    SELECT continuation.*
    INTO v_continuation
    FROM iam_initial_password_continuations AS continuation
    WHERE continuation.selector OPERATOR(pg_catalog.=) p_selector
    FOR UPDATE;
    IF NOT FOUND OR v_continuation.tenant_id IS DISTINCT FROM p_tenant_id OR
       v_continuation.state IS DISTINCT FROM 'active' OR
       v_continuation.revision IS DISTINCT FROM p_expected_revision OR
       v_continuation.secret_hmac IS DISTINCT FROM p_secret_hmac OR
       v_continuation.csrf_hmac IS DISTINCT FROM p_csrf_hmac OR
       v_continuation.policy_attempts OPERATOR(pg_catalog.>=) 5 OR
       v_now OPERATOR(pg_catalog.<) v_continuation.issued_at OR
       v_now OPERATOR(pg_catalog.>=) v_continuation.expires_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM initial password continuation changed';
    END IF;
    v_attempts := v_continuation.policy_attempts OPERATOR(pg_catalog.+) 1;
    v_state := CASE WHEN v_attempts OPERATOR(pg_catalog.=) 5
        THEN 'exhausted' ELSE 'active' END;
    UPDATE iam_initial_password_continuations AS continuation
    SET policy_attempts = v_attempts, state = v_state,
        revision = continuation.revision OPERATOR(pg_catalog.+) 1,
        updated_at = v_now,
        consumed_at = CASE WHEN v_state OPERATOR(pg_catalog.=) 'exhausted'
            THEN v_now ELSE NULL END
    WHERE continuation.selector OPERATOR(pg_catalog.=) p_selector;
    RETURN QUERY SELECT v_state, v_attempts,
        p_expected_revision OPERATOR(pg_catalog.+) 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_password_reject_v1(
    name,uuid,bytea,bytea,bytea,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_password_reject_v1(
    name,uuid,bytea,bytea,bytea,bigint
) TO forgecrud_runtime;
