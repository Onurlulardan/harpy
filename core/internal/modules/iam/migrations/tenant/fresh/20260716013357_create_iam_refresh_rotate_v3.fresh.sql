-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_rotate_v3(
    p_scope_schema name,
    p_token_id uuid, p_family_id uuid,
    p_expected_token_revision bigint, p_expected_family_revision bigint,
    p_session_id uuid, p_expected_session_revision bigint,
    p_user_id uuid, p_expected_user_revision bigint,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_access_mode_id uuid, p_expected_access_mode_revision bigint,
    p_portal_binding_id uuid, p_expected_portal_binding_revision bigint,
    p_successor_id uuid, p_successor_selector bytea,
    p_successor_hash_key_id text, p_successor_secret_hmac bytea
)
RETURNS TABLE (
    outcome text, issued_at timestamptz,
    locator_selector bytea, locator_expected_revision bigint,
    locator_revision bigint, locator_source_sha256 bytea,
    locator_state text, locator_projected_at timestamptz,
    locator_retain_until timestamptz, successor_selector bytea,
    successor_revision bigint, successor_source_sha256 bytea,
    successor_state text, successor_issued_at timestamptz,
    successor_projected_at timestamptz,
    successor_family_expires_at timestamptz,
    successor_retain_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    SELECT rotated.outcome, rotated.issued_at, rotated.locator_selector,
           rotated.locator_expected_revision, rotated.locator_revision,
           rotated.locator_source_sha256, rotated.locator_state,
           rotated.locator_projected_at, rotated.locator_retain_until,
           rotated.successor_selector, rotated.successor_revision,
           rotated.successor_source_sha256, rotated.successor_state,
           rotated.successor_issued_at, rotated.successor_projected_at,
           rotated.successor_family_expires_at,
           rotated.successor_retain_until
    INTO STRICT outcome, issued_at, locator_selector,
         locator_expected_revision, locator_revision, locator_source_sha256,
         locator_state, locator_projected_at, locator_retain_until,
         successor_selector, successor_revision, successor_source_sha256,
         successor_state, successor_issued_at, successor_projected_at,
         successor_family_expires_at, successor_retain_until
    FROM iam_refresh_rotate_v2(
        p_scope_schema, p_token_id, p_family_id,
        p_expected_token_revision, p_expected_family_revision,
        p_session_id, p_expected_session_revision,
        p_user_id, p_expected_user_revision, p_expected_token_version,
        p_expected_authorization_version, p_access_mode_id,
        p_expected_access_mode_revision, p_portal_binding_id,
        p_expected_portal_binding_revision, p_successor_id,
        p_successor_selector, p_successor_hash_key_id,
        p_successor_secret_hmac
    ) AS rotated;
    IF locator_selector IS NOT NULL THEN
        SELECT iam_refresh_locator_source_sha256_v1(
            token.id, token.family_id, token.selector,
            token.revision, token.state, token.secret_hmac
        ) INTO STRICT locator_source_sha256
        FROM iam_refresh_tokens AS token
        WHERE token.selector OPERATOR(pg_catalog.=) locator_selector
          AND token.revision OPERATOR(pg_catalog.=) locator_revision
          AND token.state OPERATOR(pg_catalog.=) locator_state;
    END IF;
    IF successor_selector IS NOT NULL THEN
        SELECT iam_refresh_locator_source_sha256_v1(
            token.id, token.family_id, token.selector,
            token.revision, token.state, token.secret_hmac
        ) INTO STRICT successor_source_sha256
        FROM iam_refresh_tokens AS token
        WHERE token.selector OPERATOR(pg_catalog.=) successor_selector
          AND token.revision OPERATOR(pg_catalog.=) successor_revision
          AND token.state OPERATOR(pg_catalog.=) successor_state;
    END IF;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_rotate_v3(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea
) FROM PUBLIC;
