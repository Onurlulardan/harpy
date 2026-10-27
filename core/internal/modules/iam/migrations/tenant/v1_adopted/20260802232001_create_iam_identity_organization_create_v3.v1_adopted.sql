-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_organization_create_v3(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_organization_id uuid, p_name text, p_slug text, p_parent_id uuid,
    p_owner_ids uuid[], p_membership_ids uuid[],
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
    IF p_organization_id IS NULL OR p_parent_id IS NOT DISTINCT FROM p_organization_id OR
       pg_catalog.char_length(p_name) NOT BETWEEN 1 AND 255 OR
       p_name <> pg_catalog.btrim(p_name) OR
       p_slug !~ '^[a-z0-9]([a-z0-9-]{0,98}[a-z0-9])?$' OR
       (p_parent_id IS NOT NULL AND NOT EXISTS (
          SELECT 1 FROM iam_organizations WHERE id = p_parent_id AND status = 'active'
       )) OR NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE =
          'IAM identity organization create input is invalid'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.organization.create',
        p_request_sha256, p_organization_id
    );
    IF FOUND THEN RETURN; END IF;
    v_values := iam_identity_profile_assert_v2(
        'organization', p_profile_schema_id, p_profile_document, p_profile_sha256
    );
    INSERT INTO iam_organizations (
        id, name, slug, status, parent_id, profile_resource_kind,
        profile_schema_version_id, profile_values, profile_sha256,
        revision, created_at, updated_at
    ) VALUES (
        p_organization_id, p_name, p_slug, 'active', p_parent_id, 'organization',
        p_profile_schema_id, v_values, p_profile_sha256, 1, v_now, v_now
    );
    PERFORM iam_identity_organization_owners_replace_v1(
        p_organization_id, 1, p_owner_ids, p_membership_ids
    );
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.organization.create', p_request_sha256,
        p_organization_id, 1, p_request_sha256
    );
    RETURN QUERY SELECT p_organization_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_organization_create_v3(
    name,uuid,bytea,uuid,text,text,uuid,uuid[],uuid[],uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) FROM PUBLIC;
