-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_organization_read_v2(
    p_scope_schema name, p_organization_id uuid,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, name text, slug text, status text, parent_id uuid,
    revision bigint, created_at timestamptz, updated_at timestamptz,
    tombstoned_at timestamptz, owner_ids uuid[]
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_organization_id IS NULL OR NOT iam_identity_query_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY
    SELECT organization.id, organization.name, organization.slug,
        organization.status, organization.parent_id, organization.revision,
        organization.created_at, organization.updated_at,
        organization.tombstoned_at,
        COALESCE((SELECT pg_catalog.array_agg(member.user_id ORDER BY member.user_id)
          FROM iam_organization_memberships AS member
          WHERE member.organization_id = organization.id
            AND member.membership_kind = 'owner'
            AND member.state = 'active'), '{}'::uuid[])
    FROM iam_organizations AS organization
    WHERE organization.id = p_organization_id;
END;
$function$;

CREATE FUNCTION iam_identity_organization_list_v2(
    p_scope_schema name, p_after_id uuid, p_limit integer,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, name text, slug text, status text, parent_id uuid,
    revision bigint, created_at timestamptz, updated_at timestamptz,
    tombstoned_at timestamptz, owner_ids uuid[]
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit NOT BETWEEN 1 AND 100 OR NOT iam_identity_query_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY
    SELECT organization.id, organization.name, organization.slug,
        organization.status, organization.parent_id, organization.revision,
        organization.created_at, organization.updated_at,
        organization.tombstoned_at,
        COALESCE((SELECT pg_catalog.array_agg(member.user_id ORDER BY member.user_id)
          FROM iam_organization_memberships AS member
          WHERE member.organization_id = organization.id
            AND member.membership_kind = 'owner'
            AND member.state = 'active'), '{}'::uuid[])
    FROM iam_organizations AS organization
    WHERE p_after_id IS NULL OR organization.id > p_after_id
    ORDER BY organization.id LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_organization_read_v2(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_identity_organization_list_v2(
    name,uuid,integer,uuid,uuid,bigint,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_identity_organization_read_v2(
    name,uuid,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_list_v2(
    name,uuid,integer,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_organization_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_organization_list_v1(
    name,uuid,integer,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
