-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_one_time_consume_v1(
    p_scope_schema name,
    p_token_id uuid,
    p_expected_revision bigint,
    p_expected_purpose text
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_purpose text;
    v_state text;
    v_revision bigint;
    v_expires_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT (
           p_expected_purpose OPERATOR(pg_catalog.=) 'password_reset_v1' OR
           p_expected_purpose OPERATOR(pg_catalog.=) 'email_verification_v1'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM one-time consume scope or purpose is invalid';
    END IF;
    SELECT token.purpose, token.state, token.revision, token.expires_at
    INTO v_purpose, v_state, v_revision, v_expires_at
    FROM iam_one_time_tokens AS token
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM one-time token is unknown';
    END IF;
    IF v_purpose IS DISTINCT FROM p_expected_purpose OR
       v_state IS DISTINCT FROM 'pending' OR
       v_revision IS DISTINCT FROM p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM one-time consume CAS is stale';
    END IF;
    IF v_now OPERATOR(pg_catalog.>=) v_expires_at THEN
        UPDATE iam_one_time_tokens
        SET state = 'expired', expired_at = v_expires_at,
            updated_at = v_now,
            revision = v_revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_token_id;
        RETURN 'expired';
    END IF;
    UPDATE iam_one_time_tokens
    SET state = 'consumed', consumed_at = v_now, updated_at = v_now,
        revision = v_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_token_id;
    RETURN 'consumed';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_one_time_consume_v1(
    name,uuid,bigint,text
) FROM PUBLIC;
