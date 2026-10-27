-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_membership_list_v1(
    p_scope_schema name, p_after_id uuid, p_limit integer,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, organization_id uuid, user_id uuid, membership_kind text,
    state text, revision bigint, created_at timestamptz,
    updated_at timestamptz, tombstoned_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit NOT BETWEEN 1 AND 100 OR NOT iam_identity_query_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY SELECT membership.id, membership.organization_id,
        membership.user_id, membership.membership_kind, membership.state,
        membership.revision, membership.created_at,
        membership.updated_at, membership.tombstoned_at
    FROM iam_organization_memberships AS membership
    WHERE p_after_id IS NULL OR membership.id > p_after_id
    ORDER BY membership.id LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_membership_list_v1(
    name,uuid,integer,uuid,uuid,bigint,bigint
) FROM PUBLIC;
