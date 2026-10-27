-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_decide_v2(
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
    p_expected_portal_binding_revision bigint
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_token_state text;
    v_token_revision bigint;
    v_token_sequence bigint;
    v_token_issued_at timestamptz;
    v_token_updated_at timestamptz;
    v_idle_expires_at timestamptz;
    v_family_state text;
    v_family_revision bigint;
    v_rotation_sequence bigint;
    v_family_expires_at timestamptz;
    v_family_updated_at timestamptz;
    v_session_state text;
    v_session_revision bigint;
    v_session_expires_at timestamptz;
    v_session_updated_at timestamptz;
    v_access_mode text;
    v_portal_binding_id uuid;
    v_user_status text;
    v_identity_kind text;
    v_user_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_user_updated_at timestamptz;
    v_access_state text;
    v_access_revision bigint;
    v_access_updated_at timestamptz;
    v_binding_state text;
    v_binding_revision bigint;
    v_binding_updated_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_token_id IS NULL OR p_family_id IS NULL OR p_session_id IS NULL OR
       p_user_id IS NULL OR p_access_mode_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh decision scope is invalid';
    END IF;
    SELECT token.state, token.revision, token.sequence, token.issued_at,
           token.updated_at, token.idle_expires_at,
           family.state, family.revision, family.rotation_sequence,
           family.absolute_expires_at, family.updated_at,
           session.state, session.revision, session.absolute_expires_at,
           session.updated_at, session.access_mode, session.portal_binding_id,
           principal.status, principal.identity_kind, principal.revision,
           principal.token_version, principal.authorization_version,
           principal.updated_at, access.state, access.revision, access.updated_at
    INTO v_token_state, v_token_revision, v_token_sequence, v_token_issued_at,
         v_token_updated_at, v_idle_expires_at,
         v_family_state, v_family_revision, v_rotation_sequence,
         v_family_expires_at, v_family_updated_at,
         v_session_state, v_session_revision, v_session_expires_at,
         v_session_updated_at, v_access_mode, v_portal_binding_id,
         v_user_status, v_identity_kind, v_user_revision,
         v_token_version, v_authorization_version, v_user_updated_at,
         v_access_state, v_access_revision, v_access_updated_at
    FROM iam_refresh_tokens AS token
    JOIN iam_refresh_families AS family
      ON family.id OPERATOR(pg_catalog.=) token.family_id
    JOIN iam_sessions AS session
      ON session.id OPERATOR(pg_catalog.=) family.session_id
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) family.user_id
    JOIN iam_user_access_modes AS access
      ON access.id OPERATOR(pg_catalog.=) session.access_mode_id
     AND access.user_id OPERATOR(pg_catalog.=) principal.id
     AND access.mode OPERATOR(pg_catalog.=) session.access_mode
    WHERE token.id OPERATOR(pg_catalog.=) p_token_id
      AND family.id OPERATOR(pg_catalog.=) p_family_id
      AND session.id OPERATOR(pg_catalog.=) p_session_id
      AND principal.id OPERATOR(pg_catalog.=) p_user_id
      AND access.id OPERATOR(pg_catalog.=) p_access_mode_id
    FOR UPDATE OF principal, access, session, family, token;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM refresh decision subject is unavailable';
    END IF;
    IF v_portal_binding_id IS NOT NULL THEN
        SELECT binding.state, binding.revision, binding.updated_at
        INTO v_binding_state, v_binding_revision, v_binding_updated_at
        FROM iam_portal_subject_bindings AS binding
        WHERE binding.id OPERATOR(pg_catalog.=) v_portal_binding_id
          AND binding.user_id OPERATOR(pg_catalog.=) p_user_id
          AND binding.access_mode_id OPERATOR(pg_catalog.=) p_access_mode_id
        FOR UPDATE;
    END IF;
    IF v_token_revision IS DISTINCT FROM p_expected_token_revision OR
       v_family_revision IS DISTINCT FROM p_expected_family_revision OR
       v_session_revision IS DISTINCT FROM p_expected_session_revision OR
       v_user_revision IS DISTINCT FROM p_expected_user_revision OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version OR
       v_access_revision IS DISTINCT FROM p_expected_access_mode_revision OR
       v_portal_binding_id IS DISTINCT FROM p_portal_binding_id OR
       v_binding_revision IS DISTINCT FROM p_expected_portal_binding_revision OR
       v_now OPERATOR(pg_catalog.<) GREATEST(
           v_token_updated_at, v_family_updated_at, v_session_updated_at,
           v_user_updated_at, v_access_updated_at, v_binding_updated_at
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM refresh decision CAS is stale';
    END IF;
    IF v_token_state IS NOT DISTINCT FROM 'consumed' THEN
        IF v_family_state IS NOT DISTINCT FROM 'active' THEN
            RETURN 'reused';
        END IF;
        RETURN 'rejected';
    END IF;
    IF v_token_state IS DISTINCT FROM 'active' OR
       v_family_state IS DISTINCT FROM 'active' OR
       v_session_state IS DISTINCT FROM 'active' OR
       v_user_status IS DISTINCT FROM 'active' OR
       v_identity_kind IS DISTINCT FROM 'human' OR
       v_access_state IS DISTINCT FROM 'active' OR NOT (
           (v_access_mode IS NOT DISTINCT FROM 'operator_console' AND
            v_portal_binding_id IS NULL) OR
           (v_access_mode IS NOT DISTINCT FROM 'portal' AND
            v_binding_state IS NOT DISTINCT FROM 'active')
       ) THEN
        RETURN 'rejected';
    END IF;
    IF v_now OPERATOR(pg_catalog.>=) v_idle_expires_at OR
       v_now OPERATOR(pg_catalog.>=) v_family_expires_at OR
       v_now OPERATOR(pg_catalog.>=) v_session_expires_at THEN
        RETURN 'expired';
    END IF;
    IF v_now OPERATOR(pg_catalog.<=) v_token_issued_at OR
       v_rotation_sequence IS DISTINCT FROM
           v_token_sequence OPERATOR(pg_catalog.-) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM refresh decision sequence is stale';
    END IF;
    RETURN 'rotate';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_decide_v2(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint
) FROM PUBLIC;
