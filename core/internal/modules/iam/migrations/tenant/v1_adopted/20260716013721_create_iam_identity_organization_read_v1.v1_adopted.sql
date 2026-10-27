-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_organization_read_v1(
    p_scope_schema name, p_organization_id uuid,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, name text, slug text, status text, parent_id uuid,
    revision bigint, created_at timestamptz, updated_at timestamptz,
    tombstoned_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_organization_id IS NULL OR NOT iam_identity_query_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY SELECT organization.id, organization.name, organization.slug,
        organization.status,
        COALESCE(organization.parent_id, '00000000-0000-0000-0000-000000000000'::uuid),
        organization.revision, organization.created_at,
        organization.updated_at, organization.tombstoned_at
    FROM iam_organizations AS organization WHERE organization.id = p_organization_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_organization_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
