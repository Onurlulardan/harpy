-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_read_v1(
    p_scope_schema name, p_user_id uuid, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, status text, token_version bigint, authorization_version bigint,
    revision bigint, identifier_id uuid, email_display text,
    email_normalized text, role_ids uuid[], organization_ids uuid[],
    created_at timestamptz, updated_at timestamptz, tombstoned_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
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
        COALESCE((SELECT pg_catalog.array_agg(assignment.role_id ORDER BY assignment.role_id)
          FROM iam_user_role_assignments AS assignment
          WHERE assignment.user_id = principal.id AND assignment.state <> 'revoked'),
          '{}'::uuid[]),
        COALESCE((SELECT pg_catalog.array_agg(membership.organization_id
                                              ORDER BY membership.organization_id)
          FROM iam_organization_memberships AS membership
          WHERE membership.user_id = principal.id AND membership.state <> 'tombstoned'),
          '{}'::uuid[]),
        principal.created_at, principal.updated_at, principal.tombstoned_at
    FROM iam_users AS principal
    JOIN LATERAL (
        SELECT source.id, source.display_value, source.normalized_value
        FROM iam_user_identifiers AS source
        WHERE source.user_id = principal.id AND source.kind = 'email_v1'
          AND source.state <> 'retired'
        ORDER BY source.revision DESC LIMIT 1
    ) AS email ON true
    WHERE principal.id = p_user_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_user_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
