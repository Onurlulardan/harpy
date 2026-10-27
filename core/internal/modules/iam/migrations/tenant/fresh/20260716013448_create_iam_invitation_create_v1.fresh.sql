-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_create_v1(
    p_scope_schema name,
    p_invitation_id uuid,
    p_display_value text,
    p_normalized_value text,
    p_normalized_sha256 bytea,
    p_role_id uuid,
    p_role_scope_kind text,
    p_role_scope_id uuid,
    p_organization_id uuid,
    p_invited_by uuid,
    p_expected_authorization_version bigint,
    p_token_id uuid,
    p_selector bytea,
    p_hash_key_id text,
    p_secret_hmac bytea,
    p_binding_sha256 bytea
)
RETURNS TABLE (invitation_id uuid, invitation_state text, invitation_revision bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_expires_at timestamptz;
    v_user_state text;
    v_authorization_version bigint;
    v_role_state text;
    v_role_scope_kind text;
    v_role_scope_id uuid;
    v_role_protected boolean;
BEGIN
    v_expires_at := v_now + INTERVAL '7 days';
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation creation scope is invalid';
    END IF;
    IF p_invitation_id IS NULL OR p_token_id IS NULL OR
       p_role_id IS NULL OR p_role_scope_id IS NULL OR p_invited_by IS NULL OR
       p_expected_authorization_version < 1 OR
       p_role_scope_kind NOT IN ('tenant', 'organization') OR
       (p_role_scope_kind = 'tenant' AND p_organization_id IS NOT NULL) OR
       (p_role_scope_kind = 'organization' AND
           p_organization_id IS DISTINCT FROM p_role_scope_id) OR
       pg_catalog.char_length(p_display_value) NOT BETWEEN 1 AND 320 OR
       pg_catalog.octet_length(p_normalized_value) NOT BETWEEN 3 AND 254 OR
       pg_catalog.octet_length(p_normalized_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       p_hash_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_binding_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM invitation creation input is invalid';
    END IF;
    SELECT principal.status, principal.authorization_version
    INTO v_user_state, v_authorization_version
    FROM iam_users AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_invited_by
    FOR UPDATE;
    IF NOT FOUND OR v_user_state IS DISTINCT FROM 'active' OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version OR
       NOT iam_invitation_manager_v1(p_invited_by) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation creation authority is denied';
    END IF;
    SELECT role.status, role.scope_kind, role.scope_id, role.is_protected
    INTO v_role_state, v_role_scope_kind, v_role_scope_id, v_role_protected
    FROM iam_roles AS role
    WHERE role.id OPERATOR(pg_catalog.=) p_role_id
    FOR UPDATE;
    IF NOT FOUND OR v_role_state IS DISTINCT FROM 'active' OR v_role_protected OR
       v_role_scope_kind IS DISTINCT FROM p_role_scope_kind OR
       v_role_scope_id IS DISTINCT FROM p_role_scope_id OR
       (p_organization_id IS NOT NULL AND NOT EXISTS (
           SELECT 1 FROM iam_organizations AS organization
           WHERE organization.id OPERATOR(pg_catalog.=) p_organization_id
             AND organization.status OPERATOR(pg_catalog.=) 'active'
       )) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation role authority is denied';
    END IF;
    INSERT INTO iam_invitations (
        id, identifier_kind, display_value, normalized_value, normalized_sha256,
        role_id, role_scope_kind, role_scope_id, organization_id, invited_by,
        accepted_user_id, accepted_identifier_id, accepted_identifier_kind,
        state, revision, created_at, updated_at, expires_at,
        accepted_at, revoked_at, expired_at
    ) VALUES (
        p_invitation_id, 'email_v1', p_display_value, p_normalized_value,
        p_normalized_sha256, p_role_id, p_role_scope_kind, p_role_scope_id,
        p_organization_id, p_invited_by, NULL, NULL, NULL,
        'pending', 1, v_now, v_now, v_expires_at, NULL, NULL, NULL
    );
    INSERT INTO iam_one_time_tokens (
        id, purpose, user_id, identifier_id, identifier_kind,
        identifier_revision, invitation_id, invitation_expires_at,
        invitation_revision, portal_binding_id, portal_setup_revision,
        selector, hash_key_id, secret_hmac, binding_sha256, state,
        attempt_count, revision, created_at, updated_at, expires_at,
        consumed_at, revoked_at, expired_at, purge_after
    ) VALUES (
        p_token_id, 'invitation_accept_v1', NULL, NULL, NULL, NULL,
        p_invitation_id, v_expires_at, 1, NULL, NULL,
        p_selector, p_hash_key_id, p_secret_hmac, p_binding_sha256,
        'pending', 0, 1, v_now, v_now, v_expires_at,
        NULL, NULL, NULL, v_expires_at + INTERVAL '24 hours'
    );
    RETURN QUERY SELECT p_invitation_id, 'pending'::text, 1::bigint;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_create_v1(
    name,uuid,text,text,bytea,uuid,text,uuid,uuid,uuid,bigint,
    uuid,bytea,text,bytea,bytea
) FROM PUBLIC;
