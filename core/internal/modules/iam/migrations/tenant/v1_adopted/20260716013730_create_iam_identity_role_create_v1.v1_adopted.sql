-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_role_create_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_role_id uuid, p_scope_kind text,
    p_scope_id uuid, p_organization_id uuid, p_name text,
    p_description text, p_is_default boolean,
    p_profile_schema_id uuid, p_profile_document bytea, p_profile_sha256 bytea,
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
    v_values jsonb; v_clock bigint;
BEGIN
    IF p_role_id IS NULL OR p_scope_kind NOT IN ('tenant', 'organization') OR
       (p_scope_kind = 'tenant' AND (p_scope_id <> p_tenant_id OR p_organization_id IS NOT NULL)) OR
       (p_scope_kind = 'organization' AND (p_scope_id IS DISTINCT FROM p_organization_id OR
          NOT EXISTS (SELECT 1 FROM iam_organizations
                      WHERE id = p_organization_id AND status = 'active'))) OR
       pg_catalog.char_length(p_name) NOT BETWEEN 1 AND 255 OR
       p_name <> pg_catalog.btrim(p_name) OR
       pg_catalog.char_length(COALESCE(p_description, '')) > 2000 OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
          MESSAGE = 'IAM identity role create input is invalid'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.role.create', p_request_sha256, p_role_id
    );
    IF FOUND THEN RETURN; END IF;
    v_values := iam_identity_profile_assert_v1(
        'role', p_profile_schema_id, p_profile_document, p_profile_sha256
    );
    INSERT INTO iam_roles (
        id, scope_kind, scope_id, organization_id, name, name_key, description,
        status, is_protected, is_default, profile_resource_kind,
        profile_schema_version_id, profile_values, profile_sha256,
        revision, created_at, updated_at
    ) VALUES (
        p_role_id, p_scope_kind, p_scope_id, p_organization_id, p_name,
        pg_catalog.convert_to(p_name, 'UTF8'), NULLIF(p_description, ''),
        'active', false, p_is_default, 'role', p_profile_schema_id,
        v_values, p_profile_sha256, 1, v_now, v_now
    );
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.role.create', p_request_sha256,
        p_role_id, 1, p_request_sha256
    );
    RETURN QUERY SELECT p_role_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_role_create_v1(
    name,uuid,uuid,bytea,uuid,text,uuid,uuid,text,text,boolean,
    uuid,bytea,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;
