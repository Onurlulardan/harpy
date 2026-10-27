-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_recovery_consume_v1(
    p_scope_schema name,
    p_recovery_code_id uuid,
    p_expected_revision bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_code_state text;
    v_code_revision bigint;
    v_authenticator_id uuid;
    v_authenticator_state text;
    v_authenticator_revision bigint;
    v_user_id uuid;
    v_user_revision bigint;
    v_token_version bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM recovery consume scope is invalid';
    END IF;
    SELECT code.state, code.revision, code.authenticator_id,
           authenticator.state, authenticator.revision, code.user_id,
           principal.revision, principal.token_version
    INTO v_code_state, v_code_revision, v_authenticator_id,
         v_authenticator_state, v_authenticator_revision, v_user_id,
         v_user_revision, v_token_version
    FROM iam_recovery_codes AS code
    JOIN iam_authenticators AS authenticator
      ON authenticator.id OPERATOR(pg_catalog.=) code.authenticator_id
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) code.user_id
    WHERE code.id OPERATOR(pg_catalog.=) p_recovery_code_id
    FOR UPDATE OF principal, authenticator, code;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM recovery code is unknown';
    END IF;
    IF v_code_state IS DISTINCT FROM 'active' OR
       v_code_revision IS DISTINCT FROM p_expected_revision OR
       v_authenticator_state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM recovery consume CAS is stale';
    END IF;
    UPDATE iam_recovery_codes
    SET state = 'consumed', consumed_at = v_now,
        purge_after = v_now OPERATOR(pg_catalog.+) INTERVAL '30 days',
        updated_at = v_now,
        revision = v_code_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_recovery_code_id;
    UPDATE iam_authenticators
    SET state = 'renewal_required', renewal_required_at = v_now,
        updated_at = v_now,
        revision = v_authenticator_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) v_authenticator_id;
    UPDATE iam_users
    SET token_version = v_token_version OPERATOR(pg_catalog.+) 1,
        updated_at = v_now,
        revision = v_user_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) v_user_id;
    RETURN v_token_version OPERATOR(pg_catalog.+) 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_recovery_consume_v1(
    name,uuid,bigint
) FROM PUBLIC;
