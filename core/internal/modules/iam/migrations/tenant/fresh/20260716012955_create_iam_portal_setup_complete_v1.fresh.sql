-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_setup_complete_v1(
    p_scope_schema name,
    p_binding_id uuid,
    p_expected_setup_revision bigint,
    p_token_id uuid,
    p_expected_token_revision bigint
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_setup_state text;
    v_setup_revision bigint;
    v_user_id uuid;
    v_token_state text;
    v_token_revision bigint;
    v_token_user_id uuid;
    v_token_binding_id uuid;
    v_token_setup_revision bigint;
    v_token_expires_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM portal setup completion scope is invalid';
    END IF;
    SELECT setup.setup_state, setup.revision, setup.user_id
    INTO v_setup_state, v_setup_revision, v_user_id
    FROM iam_portal_access_setups AS setup
    WHERE setup.binding_id OPERATOR(pg_catalog.=) p_binding_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM portal setup is unknown';
    END IF;
    SELECT token.state, token.revision, token.user_id,
           token.portal_binding_id, token.portal_setup_revision,
           token.expires_at
    INTO v_token_state, v_token_revision, v_token_user_id,
         v_token_binding_id, v_token_setup_revision, v_token_expires_at
    FROM iam_one_time_tokens AS token
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
      AND token.purpose OPERATOR(pg_catalog.=) 'portal_setup_v1'
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM portal setup token is unknown';
    END IF;
    IF v_setup_state IS DISTINCT FROM 'required' OR
       v_setup_revision IS DISTINCT FROM p_expected_setup_revision OR
       v_token_state IS DISTINCT FROM 'pending' OR
       v_token_revision IS DISTINCT FROM p_expected_token_revision OR
       v_token_user_id IS DISTINCT FROM v_user_id OR
       v_token_binding_id IS DISTINCT FROM p_binding_id OR
       v_token_setup_revision IS DISTINCT FROM p_expected_setup_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM portal setup completion CAS is stale';
    END IF;
    IF v_now OPERATOR(pg_catalog.>=) v_token_expires_at THEN
        UPDATE iam_one_time_tokens
        SET state = 'expired', expired_at = v_token_expires_at,
            updated_at = v_now,
            revision = v_token_revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_token_id;
        RETURN 'expired';
    END IF;
    UPDATE iam_portal_access_setups
    SET setup_state = 'completed', setup_completed_at = v_now,
        updated_at = v_now,
        revision = v_setup_revision OPERATOR(pg_catalog.+) 1
    WHERE binding_id OPERATOR(pg_catalog.=) p_binding_id;
    UPDATE iam_one_time_tokens
    SET state = 'consumed', consumed_at = v_now, updated_at = v_now,
        revision = v_token_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_token_id;
    RETURN 'completed';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_setup_complete_v1(
    name,uuid,bigint,uuid,bigint
) FROM PUBLIC;
