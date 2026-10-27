-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_assignment_read_v1(
    p_scope_schema name, p_assignment_id uuid,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid, user_id uuid, role_id uuid, role_scope_kind text, role_scope_id uuid,
    organization_id uuid, membership_id uuid, state text, revision bigint,
    created_at timestamptz, updated_at timestamptz, revoked_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_assignment_id IS NULL OR NOT iam_identity_query_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY SELECT assignment.id, assignment.user_id, assignment.role_id,
        assignment.role_scope_kind, assignment.role_scope_id,
        COALESCE(assignment.organization_id, '00000000-0000-0000-0000-000000000000'::uuid),
        COALESCE(assignment.membership_id, '00000000-0000-0000-0000-000000000000'::uuid),
        assignment.state, assignment.revision, assignment.created_at,
        assignment.updated_at, assignment.revoked_at
    FROM iam_user_role_assignments AS assignment WHERE assignment.id = p_assignment_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_assignment_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
