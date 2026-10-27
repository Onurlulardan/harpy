-- Recovery consumption boundary.
-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_recovery_consume_v2(
    p_scope_schema name, p_scope_binding bytea,
    p_hmac_key_id text, p_code_hmac bytea, p_code_id uuid,
    p_expected_code_revision bigint, p_authenticator_id uuid,
    p_expected_authenticator_revision bigint, p_user_id uuid,
    p_expected_user_revision bigint, p_expected_token_version bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_code_revision bigint;
    v_authenticator_revision bigint;
    v_user_revision bigint;
    v_token_version bigint;
    v_code_state text;
    v_authenticator_state text;
    v_user_status text;
    v_identity_kind text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_scope_binding) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_code_hmac) IS DISTINCT FROM 32 OR
       p_scope_binding IS DISTINCT FROM p_code_hmac THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM recovery consume scope is invalid';
    END IF;
    SELECT code.revision, authenticator.revision, principal.revision,
           principal.token_version, code.state, authenticator.state,
           principal.status, principal.identity_kind
    INTO v_code_revision, v_authenticator_revision, v_user_revision,
         v_token_version, v_code_state, v_authenticator_state,
         v_user_status, v_identity_kind
    FROM iam_recovery_codes AS code
    JOIN iam_authenticators AS authenticator
      ON authenticator.id OPERATOR(pg_catalog.=) code.authenticator_id
     AND authenticator.user_id OPERATOR(pg_catalog.=) code.user_id
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) code.user_id
    WHERE code.id OPERATOR(pg_catalog.=) p_code_id
      AND code.authenticator_id OPERATOR(pg_catalog.=) p_authenticator_id
      AND code.user_id OPERATOR(pg_catalog.=) p_user_id
      AND code.hmac_key_id OPERATOR(pg_catalog.=) p_hmac_key_id
      AND code.code_hmac OPERATOR(pg_catalog.=) p_code_hmac
    FOR UPDATE OF code, authenticator, principal;
    IF NOT FOUND OR v_code_revision IS DISTINCT FROM
           p_expected_code_revision OR
       v_authenticator_revision IS DISTINCT FROM
           p_expected_authenticator_revision OR
       v_user_revision IS DISTINCT FROM p_expected_user_revision OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_code_state IS DISTINCT FROM 'active' OR
       v_authenticator_state IS DISTINCT FROM 'active' OR
       v_user_status IS DISTINCT FROM 'active' OR
       v_identity_kind IS DISTINCT FROM 'human' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM recovery consume CAS is stale';
    END IF;
    RETURN iam_recovery_consume_v1(
        p_scope_schema, p_code_id, p_expected_code_revision
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_recovery_consume_v2(
    name,bytea,text,bytea,uuid,bigint,uuid,bigint,uuid,bigint,bigint
) FROM PUBLIC;
