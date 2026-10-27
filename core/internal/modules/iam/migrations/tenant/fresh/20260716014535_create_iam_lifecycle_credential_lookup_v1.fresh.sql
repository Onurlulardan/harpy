-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_lifecycle_credential_lookup_v1(
    p_scope_schema name, p_purpose text, p_selector bytea,
    p_binding_sha256 bytea
)
RETURNS TABLE (
    token_id uuid, token_purpose text, user_id uuid, identifier_id uuid,
    identifier_kind text, identifier_revision bigint,
    portal_binding_id uuid, portal_setup_revision bigint,
    selector bytea, hash_key_id text, secret_hmac bytea,
    binding_sha256 bytea, token_revision bigint, token_state text,
    expires_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_purpose NOT IN (
           'email_verification_v1', 'password_reset_v1', 'portal_setup_v1'
       ) OR pg_catalog.octet_length(p_selector) <> 16 OR
       pg_catalog.octet_length(p_binding_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM lifecycle credential lookup is denied';
    END IF;
    RETURN QUERY
    SELECT token.id, token.purpose, token.user_id, token.identifier_id,
           token.identifier_kind, token.identifier_revision,
           token.portal_binding_id, token.portal_setup_revision,
           token.selector, token.hash_key_id, token.secret_hmac,
           token.binding_sha256, token.revision, token.state,
           token.expires_at
    FROM iam_one_time_tokens AS token
    WHERE token.purpose = p_purpose
      AND token.selector = p_selector
      AND token.binding_sha256 = p_binding_sha256
      AND token.state IN ('pending', 'consumed');
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM lifecycle credential is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_lifecycle_credential_lookup_v1(
    name,text,bytea,bytea
) FROM PUBLIC;
