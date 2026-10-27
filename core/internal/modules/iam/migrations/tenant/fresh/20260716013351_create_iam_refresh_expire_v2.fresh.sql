-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_expire_v2(
    p_scope_schema name,
    p_token_id uuid,
    p_family_id uuid,
    p_session_id uuid
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
    v_selector bytea;
    v_secret_hmac bytea;
    v_revision bigint;
    v_idle_expires_at timestamptz;
    v_updated_at timestamptz;
    v_retain_until timestamptz;
    v_family_expires_at timestamptz;
    v_session_expires_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh expiry scope is invalid';
    END IF;
    SELECT token.selector, token.secret_hmac, token.revision,
           token.idle_expires_at, token.updated_at, token.retain_until,
           family.absolute_expires_at, session.absolute_expires_at
    INTO v_selector, v_secret_hmac, v_revision, v_idle_expires_at,
         v_updated_at, v_retain_until, v_family_expires_at,
         v_session_expires_at
    FROM iam_refresh_tokens AS token
    JOIN iam_refresh_families AS family
      ON family.id OPERATOR(pg_catalog.=) token.family_id
    JOIN iam_sessions AS session
      ON session.id OPERATOR(pg_catalog.=) family.session_id
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
      AND family.id OPERATOR(pg_catalog.=) p_family_id
      AND session.id OPERATOR(pg_catalog.=) p_session_id;
    IF NOT FOUND OR v_now OPERATOR(pg_catalog.<) v_idle_expires_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM refresh expiry state is stale';
    END IF;
    UPDATE iam_refresh_tokens
    SET state = 'expired', expired_at = v_idle_expires_at,
        updated_at = GREATEST(v_updated_at, v_now),
        revision = v_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_token_id
      AND state OPERATOR(pg_catalog.=) 'active';
    locator_selector := v_selector;
    locator_expected_revision := v_revision;
    locator_revision := v_revision OPERATOR(pg_catalog.+) 1;
    locator_source_sha256 := v_secret_hmac;
    locator_state := 'expired';
    locator_projected_at := GREATEST(v_updated_at, v_now);
    locator_retain_until := v_retain_until;
    IF v_now OPERATOR(pg_catalog.>=) v_family_expires_at THEN
        UPDATE iam_refresh_families
        SET state = 'expired', expired_at = absolute_expires_at,
            updated_at = GREATEST(updated_at, v_now),
            revision = revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_family_id
          AND state OPERATOR(pg_catalog.=) 'active';
    END IF;
    IF v_now OPERATOR(pg_catalog.>=) v_session_expires_at THEN
        UPDATE iam_sessions
        SET state = 'expired', expired_at = absolute_expires_at,
            updated_at = GREATEST(updated_at, v_now),
            revision = revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_session_id
          AND state OPERATOR(pg_catalog.=) 'active';
    END IF;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_expire_v2(
    name,uuid,uuid,uuid
) FROM PUBLIC;
