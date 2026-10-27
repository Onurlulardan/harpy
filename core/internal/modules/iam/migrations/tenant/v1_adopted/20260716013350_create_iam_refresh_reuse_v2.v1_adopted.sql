-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_reuse_v2(
    p_scope_schema name,
    p_family_id uuid,
    p_session_id uuid,
    p_user_id uuid
)
RETURNS TABLE (
    locator_selector bytea,
    locator_expected_revision bigint,
    locator_revision bigint,
    locator_source_sha256 bytea,
    locator_state text,
    locator_projected_at timestamptz,
    locator_retain_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_successor_id uuid;
    v_selector bytea;
    v_secret_hmac bytea;
    v_revision bigint;
    v_updated_at timestamptz;
    v_retain_until timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh reuse scope is invalid';
    END IF;
    SELECT token.id, token.selector, token.secret_hmac, token.revision,
           token.updated_at, token.retain_until
    INTO v_successor_id, v_selector, v_secret_hmac, v_revision,
         v_updated_at, v_retain_until
    FROM iam_refresh_tokens AS token
    WHERE token.family_id OPERATOR(pg_catalog.=) p_family_id
      AND token.state OPERATOR(pg_catalog.=) 'active'
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM refresh reuse successor is unavailable';
    END IF;
    UPDATE iam_refresh_tokens
        SET state = 'revoked',
            revoked_at = GREATEST(v_updated_at, v_now),
            updated_at = GREATEST(v_updated_at, v_now),
            revision = v_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) v_successor_id;
    locator_selector := v_selector;
    locator_expected_revision := v_revision;
    locator_revision := v_revision OPERATOR(pg_catalog.+) 1;
    locator_source_sha256 := v_secret_hmac;
    locator_state := 'revoked';
    locator_projected_at := GREATEST(v_updated_at, v_now);
    locator_retain_until := v_retain_until;
    UPDATE iam_refresh_families
    SET state = 'compromised', revoke_reason = 'reuse_detected',
        compromised_at = GREATEST(updated_at, v_now),
        updated_at = GREATEST(updated_at, v_now),
        revision = revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_family_id
      AND user_id OPERATOR(pg_catalog.=) p_user_id
      AND session_id OPERATOR(pg_catalog.=) p_session_id
      AND state OPERATOR(pg_catalog.=) 'active';
    UPDATE iam_sessions
    SET state = 'revoked',
        revoked_at = GREATEST(updated_at, v_now),
        updated_at = GREATEST(updated_at, v_now),
        revision = revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_session_id
      AND user_id OPERATOR(pg_catalog.=) p_user_id
      AND state OPERATOR(pg_catalog.=) 'active';
    UPDATE iam_users
    SET token_version = token_version OPERATOR(pg_catalog.+) 1,
        updated_at = GREATEST(updated_at, v_now),
        revision = revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_user_id;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_reuse_v2(
    name,uuid,uuid,uuid
) FROM PUBLIC;
