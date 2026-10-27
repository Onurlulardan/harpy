-- Session revocation boundary.
-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_session_revoke_v1(
    p_scope_schema name, p_token_id uuid, p_family_id uuid,
    p_expected_token_revision bigint, p_expected_family_revision bigint,
    p_session_id uuid, p_expected_session_revision bigint,
    p_user_id uuid, p_expected_user_revision bigint
)
RETURNS TABLE (
    outcome text, locator_selector bytea,
    locator_expected_revision bigint, locator_revision bigint,
    locator_source_sha256 bytea, locator_state text,
    locator_projected_at timestamptz, locator_retain_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_token_state text;
    v_token_revision bigint;
    v_family_state text;
    v_family_revision bigint;
    v_session_state text;
    v_session_revision bigint;
    v_user_revision bigint;
    v_secret_hmac bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_token_id IS NULL OR p_family_id IS NULL OR p_session_id IS NULL OR
       p_user_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM session revoke scope is invalid';
    END IF;
    SELECT token.state, token.revision, token.selector, token.secret_hmac,
           token.retain_until, family.state, family.revision,
           session.state, session.revision, principal.revision
    INTO v_token_state, v_token_revision, locator_selector, v_secret_hmac,
         locator_retain_until, v_family_state, v_family_revision,
         v_session_state, v_session_revision, v_user_revision
    FROM iam_refresh_tokens AS token
    JOIN iam_refresh_families AS family
      ON family.id OPERATOR(pg_catalog.=) token.family_id
    JOIN iam_sessions AS session
      ON session.id OPERATOR(pg_catalog.=) family.session_id
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) family.user_id
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
      AND token.family_id OPERATOR(pg_catalog.=) p_family_id
      AND family.id OPERATOR(pg_catalog.=) p_family_id
      AND session.id OPERATOR(pg_catalog.=) p_session_id
      AND principal.id OPERATOR(pg_catalog.=) p_user_id
    FOR UPDATE OF token, family, session, principal;
    IF NOT FOUND OR v_token_revision IS DISTINCT FROM
           p_expected_token_revision OR
       v_family_revision IS DISTINCT FROM p_expected_family_revision OR
       v_session_revision IS DISTINCT FROM p_expected_session_revision OR
       v_user_revision IS DISTINCT FROM p_expected_user_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM session revoke CAS is stale';
    END IF;
    IF v_token_state IS DISTINCT FROM 'active' THEN
        outcome := 'terminal';
        locator_selector := NULL;
        locator_retain_until := NULL;
        RETURN NEXT;
        RETURN;
    END IF;
    IF v_family_state IS DISTINCT FROM 'active' OR
       v_session_state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM session revoke family is stale';
    END IF;
    UPDATE iam_refresh_tokens
    SET state = 'revoked', revoked_at = v_now, updated_at = v_now,
        revision = v_token_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_token_id;
    UPDATE iam_refresh_families
    SET state = 'revoked', revoke_reason = 'logout', revoked_at = v_now,
        updated_at = v_now,
        revision = v_family_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_family_id;
    UPDATE iam_sessions
    SET state = 'revoked', revoked_at = v_now, updated_at = v_now,
        revision = v_session_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_session_id;
    outcome := 'revoked';
    locator_expected_revision := v_token_revision;
    locator_revision := v_token_revision OPERATOR(pg_catalog.+) 1;
    locator_state := 'revoked';
    locator_projected_at := v_now;
    locator_source_sha256 := iam_refresh_locator_source_sha256_v1(
        p_token_id, p_family_id, locator_selector, locator_revision,
        locator_state, v_secret_hmac
    );
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_session_revoke_v1(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint
) FROM PUBLIC;
