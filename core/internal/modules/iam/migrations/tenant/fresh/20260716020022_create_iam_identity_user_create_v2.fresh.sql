-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_create_v2(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_user_id uuid, p_identifier_id uuid,
    p_access_mode_id uuid, p_assignment_id uuid,
    p_profile_schema_id uuid, p_profile_document bytea, p_profile_sha256 bytea,
    p_email_display text, p_email_normalized text, p_email_sha256 bytea,
    p_lookup_key_id text, p_identifier_hmac bytea,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, result_revision bigint,
    authorization_clock bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_values jsonb; v_role_id uuid; v_clock bigint;
BEGIN
    IF p_tenant_id IS NULL OR p_user_id IS NULL OR p_identifier_id IS NULL OR
       p_access_mode_id IS NULL OR p_assignment_id IS NULL OR
       p_email_display = '' OR pg_catalog.octet_length(p_email_normalized) NOT BETWEEN 3 AND 254 OR
       pg_catalog.octet_length(p_email_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.sha256(pg_catalog.convert_to(p_email_normalized, 'UTF8'))
          IS DISTINCT FROM p_email_sha256 OR
       p_lookup_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_identifier_hmac) IS DISTINCT FROM 32 OR
       NOT EXISTS (
          SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema::text
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM identity user create input is invalid';
    END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.user.create', p_request_sha256, p_user_id
    );
    IF FOUND THEN RETURN; END IF;
    v_values := iam_identity_profile_assert_v2(
        'user', p_profile_schema_id, p_profile_document, p_profile_sha256
    );
    SELECT role.id INTO v_role_id FROM iam_roles AS role
    WHERE role.scope_kind = 'tenant' AND role.scope_id = p_tenant_id
      AND role.status = 'active' AND role.is_default FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM identity default tenant role is unavailable';
    END IF;
    INSERT INTO iam_users (
        id, identity_kind, status, profile_resource_kind,
        profile_schema_version_id, profile_values, profile_sha256,
        token_version, authorization_version, revision, created_at, updated_at
    ) VALUES (
        p_user_id, 'human', 'active', 'user', p_profile_schema_id,
        v_values, p_profile_sha256, 1, 1, 1, v_now, v_now
    );
    INSERT INTO iam_user_identifiers (
        id, user_id, kind, display_value, normalized_value, normalized_sha256,
        state, revision, verified_at, activated_at, created_at, updated_at
    ) VALUES (
        p_identifier_id, p_user_id, 'email_v1', p_email_display,
        p_email_normalized, p_email_sha256, 'active', 1, v_now, v_now, v_now, v_now
    );
    INSERT INTO iam_user_access_modes (
        id, user_id, mode, state, revision, created_at, updated_at, activated_at
    ) VALUES (
        p_access_mode_id, p_user_id, 'operator_console', 'active', 1, v_now, v_now, v_now
    );
    INSERT INTO iam_user_role_assignments (
        id, user_id, role_id, role_scope_kind, role_scope_id,
        state, revision, created_at, updated_at
    ) VALUES (
        p_assignment_id, p_user_id, v_role_id, 'tenant', p_tenant_id,
        'active', 1, v_now, v_now
    );
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.user.create', p_request_sha256,
        p_user_id, 1, p_request_sha256
    );
    RETURN QUERY SELECT p_user_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_user_create_v2(
    name,uuid,uuid,bytea,uuid,uuid,uuid,uuid,uuid,bytea,bytea,
    text,text,bytea,text,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;
