-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_read_v3(
    p_scope_schema name, p_user_id uuid, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, status text, token_version bigint, authorization_version bigint,
    revision bigint, identifier_id uuid, email_display text,
    email_normalized text, role_ids uuid[], organization_ids uuid[],
    created_at timestamptz, updated_at timestamptz, tombstoned_at timestamptz,
    email_verified boolean, access_modes text[]
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_user_id IS NULL OR NOT iam_identity_query_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY
    SELECT principal.id, principal.status, principal.token_version,
        principal.authorization_version, principal.revision,
        email.id, email.display_value, email.normalized_value,
        COALESCE((SELECT pg_catalog.array_agg(role.role_id ORDER BY role.role_id)
          FROM iam_user_role_assignments AS role
          WHERE role.user_id = principal.id AND role.state = 'active'), '{}'::uuid[]),
        COALESCE((SELECT pg_catalog.array_agg(member.organization_id
                                              ORDER BY member.organization_id)
          FROM iam_organization_memberships AS member
          WHERE member.user_id = principal.id AND member.state = 'active'), '{}'::uuid[]),
        principal.created_at, principal.updated_at, principal.tombstoned_at,
        email.verified_at IS NOT NULL,
        COALESCE((SELECT pg_catalog.array_agg(access.mode ORDER BY access.mode)
          FROM iam_user_access_modes AS access
          WHERE access.user_id = principal.id AND access.state = 'active'), '{}'::text[])
    FROM iam_users AS principal
    JOIN LATERAL (
        SELECT source.id, source.display_value, source.normalized_value,
               source.verified_at
        FROM iam_user_identifiers AS source
        WHERE source.user_id = principal.id AND source.kind = 'email_v1'
          AND source.state = 'active'
        ORDER BY source.revision DESC LIMIT 1
    ) AS email ON true
    WHERE principal.id = p_user_id;
END;
$function$;

CREATE FUNCTION iam_identity_user_list_v3(
    p_scope_schema name, p_after_id uuid, p_limit integer,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, status text, token_version bigint, authorization_version bigint,
    revision bigint, identifier_id uuid, email_display text,
    email_normalized text, role_ids uuid[], organization_ids uuid[],
    created_at timestamptz, updated_at timestamptz, tombstoned_at timestamptz,
    email_verified boolean, access_modes text[]
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit NOT BETWEEN 1 AND 100 OR NOT iam_identity_query_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY
    SELECT principal.id, principal.status, principal.token_version,
        principal.authorization_version, principal.revision,
        email.id, email.display_value, email.normalized_value,
        COALESCE((SELECT pg_catalog.array_agg(role.role_id ORDER BY role.role_id)
          FROM iam_user_role_assignments AS role
          WHERE role.user_id = principal.id AND role.state = 'active'), '{}'::uuid[]),
        COALESCE((SELECT pg_catalog.array_agg(member.organization_id
                                              ORDER BY member.organization_id)
          FROM iam_organization_memberships AS member
          WHERE member.user_id = principal.id AND member.state = 'active'), '{}'::uuid[]),
        principal.created_at, principal.updated_at, principal.tombstoned_at,
        email.verified_at IS NOT NULL,
        COALESCE((SELECT pg_catalog.array_agg(access.mode ORDER BY access.mode)
          FROM iam_user_access_modes AS access
          WHERE access.user_id = principal.id AND access.state = 'active'), '{}'::text[])
    FROM iam_users AS principal
    JOIN LATERAL (
        SELECT source.id, source.display_value, source.normalized_value,
               source.verified_at
        FROM iam_user_identifiers AS source
        WHERE source.user_id = principal.id AND source.kind = 'email_v1'
          AND source.state = 'active'
        ORDER BY source.revision DESC LIMIT 1
    ) AS email ON true
    WHERE p_after_id IS NULL OR principal.id > p_after_id
    ORDER BY principal.id LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_user_read_v3(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_identity_user_list_v3(
    name,uuid,integer,uuid,uuid,bigint,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_identity_user_read_v3(
    name,uuid,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_user_list_v3(
    name,uuid,integer,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_user_read_v2(
    name,uuid,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_user_list_v2(
    name,uuid,integer,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
