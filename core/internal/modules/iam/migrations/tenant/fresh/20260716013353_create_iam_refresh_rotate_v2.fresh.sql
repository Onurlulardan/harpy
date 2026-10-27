-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_rotate_v2(
    p_scope_schema name,
    p_token_id uuid,
    p_family_id uuid,
    p_expected_token_revision bigint,
    p_expected_family_revision bigint,
    p_session_id uuid,
    p_expected_session_revision bigint,
    p_user_id uuid,
    p_expected_user_revision bigint,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_access_mode_id uuid,
    p_expected_access_mode_revision bigint,
    p_portal_binding_id uuid,
    p_expected_portal_binding_revision bigint,
    p_successor_id uuid,
    p_successor_selector bytea,
    p_successor_hash_key_id text,
    p_successor_secret_hmac bytea
)
RETURNS TABLE (
    outcome text,
    issued_at timestamptz,
    locator_selector bytea,
    locator_expected_revision bigint,
    locator_revision bigint,
    locator_source_sha256 bytea,
    locator_state text,
    locator_projected_at timestamptz,
    locator_retain_until timestamptz,
    successor_selector bytea,
    successor_revision bigint,
    successor_source_sha256 bytea,
    successor_state text,
    successor_issued_at timestamptz,
    successor_projected_at timestamptz,
    successor_family_expires_at timestamptz,
    successor_retain_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_decision text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_successor_id IS NULL OR
       pg_catalog.octet_length(p_successor_selector) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_successor_secret_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh rotation scope is invalid';
    END IF;
    v_decision := iam_refresh_decide_v2(
        p_scope_schema, p_token_id, p_family_id,
        p_expected_token_revision, p_expected_family_revision,
        p_session_id, p_expected_session_revision,
        p_user_id, p_expected_user_revision, p_expected_token_version,
        p_expected_authorization_version, p_access_mode_id,
        p_expected_access_mode_revision, p_portal_binding_id,
        p_expected_portal_binding_revision
    );
    IF v_decision IS NOT DISTINCT FROM 'reused' THEN
        SELECT reuse.locator_selector, reuse.locator_expected_revision,
               reuse.locator_revision, reuse.locator_source_sha256,
               reuse.locator_state, reuse.locator_projected_at,
               reuse.locator_retain_until
        INTO locator_selector, locator_expected_revision, locator_revision,
             locator_source_sha256, locator_state, locator_projected_at,
             locator_retain_until
        FROM iam_refresh_reuse_v2(
            p_scope_schema, p_family_id, p_session_id, p_user_id
        ) AS reuse;
        outcome := 'reused';
        RETURN NEXT;
        RETURN;
    END IF;
    IF v_decision IS NOT DISTINCT FROM 'expired' THEN
        SELECT expiry.locator_selector, expiry.locator_expected_revision,
               expiry.locator_revision, expiry.locator_source_sha256,
               expiry.locator_state, expiry.locator_projected_at,
               expiry.locator_retain_until
        INTO locator_selector, locator_expected_revision, locator_revision,
             locator_source_sha256, locator_state, locator_projected_at,
             locator_retain_until
        FROM iam_refresh_expire_v2(
            p_scope_schema, p_token_id, p_family_id, p_session_id
        ) AS expiry;
        outcome := 'expired';
        RETURN NEXT;
        RETURN;
    END IF;
    IF v_decision IS DISTINCT FROM 'rotate' THEN
        outcome := 'rejected';
        RETURN NEXT;
        RETURN;
    END IF;
    SELECT active.issued_at, active.locator_selector,
           active.locator_expected_revision, active.locator_revision,
           active.locator_source_sha256, active.locator_state,
           active.locator_projected_at, active.locator_retain_until,
           active.successor_selector, active.successor_revision,
           active.successor_source_sha256, active.successor_state,
           active.successor_issued_at, active.successor_projected_at,
           active.successor_family_expires_at, active.successor_retain_until
    INTO issued_at, locator_selector, locator_expected_revision,
         locator_revision, locator_source_sha256, locator_state,
         locator_projected_at, locator_retain_until, successor_selector,
         successor_revision, successor_source_sha256, successor_state,
         successor_issued_at, successor_projected_at,
         successor_family_expires_at, successor_retain_until
    FROM iam_refresh_rotate_active_v2(
        p_scope_schema, p_token_id, p_family_id, p_session_id,
        p_successor_id, p_successor_selector, p_successor_hash_key_id,
        p_successor_secret_hmac
    ) AS active;
    outcome := 'rotated';
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_rotate_v2(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea
) FROM PUBLIC;
