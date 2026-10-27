-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_assignment_grant_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_assignment_id uuid, p_user_id uuid, p_role_id uuid,
    p_scope_kind text, p_scope_id uuid, p_organization_id uuid, p_membership_id uuid,
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
    v_clock bigint; v_user_revision bigint;
BEGIN
    IF p_assignment_id IS NULL OR p_user_id IS NULL OR p_role_id IS NULL OR
       p_scope_kind NOT IN ('tenant', 'organization') OR
       (p_scope_kind = 'tenant' AND
          (p_organization_id IS NOT NULL OR p_membership_id IS NOT NULL)) OR
       (p_scope_kind = 'organization' AND
          (p_organization_id IS DISTINCT FROM p_scope_id OR p_membership_id IS NULL)) OR
       NOT EXISTS (SELECT 1 FROM iam_users WHERE id = p_user_id AND status = 'active') OR
       NOT EXISTS (SELECT 1 FROM iam_roles WHERE id = p_role_id
                   AND scope_kind = p_scope_kind AND scope_id = p_scope_id
                   AND status = 'active' AND NOT is_protected) OR
       (p_scope_kind = 'organization' AND NOT EXISTS (
          SELECT 1 FROM iam_organization_memberships AS membership
          JOIN iam_organizations AS organization ON organization.id = membership.organization_id
          WHERE membership.id = p_membership_id
            AND membership.organization_id = p_organization_id
            AND membership.user_id = p_user_id AND membership.state = 'active'
            AND organization.status = 'active'
       )) OR NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
          MESSAGE = 'IAM identity assignment grant input is invalid'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.assignment.grant',
        p_request_sha256, p_assignment_id
    );
    IF FOUND THEN RETURN; END IF;
    INSERT INTO iam_user_role_assignments (
        id, user_id, role_id, role_scope_kind, role_scope_id,
        organization_id, membership_id, state, revision, created_at, updated_at
    ) VALUES (
        p_assignment_id, p_user_id, p_role_id, p_scope_kind, p_scope_id,
        p_organization_id, p_membership_id, 'active', 1, v_now, v_now
    );
    v_user_revision := iam_identity_user_bump_v1(p_user_id, false);
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.assignment.grant', p_request_sha256,
        p_assignment_id, 1, p_request_sha256
    );
    RETURN QUERY SELECT p_assignment_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_assignment_grant_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,uuid,uuid,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
