-- Recovery credential lookup boundary.
-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_recovery_credential_v1(
    p_scope_schema name, p_scope_binding bytea,
    p_hmac_key_id text, p_code_hmac bytea
)
RETURNS TABLE (
    code_id uuid, authenticator_id uuid, user_id uuid,
    code_revision bigint, authenticator_revision bigint,
    user_revision bigint, token_version bigint, code_state text,
    authenticator_state text, user_status text, identity_kind text,
    hmac_key_id text, code_hmac bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_rows bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_scope_binding) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_code_hmac) IS DISTINCT FROM 32 OR
       p_scope_binding IS DISTINCT FROM p_code_hmac OR
       p_hmac_key_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM recovery credential scope is invalid';
    END IF;
    RETURN QUERY
    SELECT code.id, code.authenticator_id, code.user_id, code.revision,
           authenticator.revision, principal.revision,
           principal.token_version, code.state, authenticator.state,
           principal.status, principal.identity_kind, code.hmac_key_id,
           code.code_hmac
    FROM iam_recovery_codes AS code
    JOIN iam_authenticators AS authenticator
      ON authenticator.id OPERATOR(pg_catalog.=) code.authenticator_id
     AND authenticator.user_id OPERATOR(pg_catalog.=) code.user_id
     AND authenticator.kind OPERATOR(pg_catalog.=) 'recovery_set_v1'
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) code.user_id
    WHERE code.hmac_key_id OPERATOR(pg_catalog.=) p_hmac_key_id
      AND code.code_hmac OPERATOR(pg_catalog.=) p_code_hmac;
    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows OPERATOR(pg_catalog.>) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '21000',
            MESSAGE = 'IAM recovery credential is ambiguous';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_recovery_credential_v1(
    name,bytea,text,bytea
) FROM PUBLIC;
