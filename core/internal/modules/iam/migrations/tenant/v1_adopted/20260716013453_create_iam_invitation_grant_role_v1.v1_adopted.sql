-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_grant_role_v1(
    p_user_id uuid,
    p_role_assignment_id uuid,
    p_membership_id uuid,
    p_role_id uuid,
    p_role_scope_kind text,
    p_role_scope_id uuid,
    p_organization_id uuid,
    p_now timestamptz
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_role_state text;
    v_role_protected boolean;
    v_membership_id uuid;
    v_membership_state text;
BEGIN
    SELECT role.status, role.is_protected
    INTO v_role_state, v_role_protected
    FROM iam_roles AS role
    WHERE role.id = p_role_id
      AND role.scope_kind = p_role_scope_kind
      AND role.scope_id = p_role_scope_id
    FOR UPDATE;
    IF NOT FOUND OR v_role_state IS DISTINCT FROM 'active' OR v_role_protected THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation role grant is denied';
    END IF;
    IF p_role_scope_kind = 'tenant' THEN
        IF p_organization_id IS NOT NULL OR p_membership_id IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'IAM tenant invitation membership is invalid';
        END IF;
    ELSIF p_role_scope_kind = 'organization' THEN
        IF p_organization_id IS DISTINCT FROM p_role_scope_id OR NOT EXISTS (
            SELECT 1 FROM iam_organizations AS organization
            WHERE organization.id = p_organization_id
              AND organization.status = 'active'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'IAM invitation organization is denied';
        END IF;
        SELECT membership.id, membership.state
        INTO v_membership_id, v_membership_state
        FROM iam_organization_memberships AS membership
        WHERE membership.organization_id = p_organization_id
          AND membership.user_id = p_user_id
        FOR UPDATE;
        IF FOUND THEN
            IF v_membership_state IS DISTINCT FROM 'active' OR
               (p_membership_id IS NOT NULL AND
                   p_membership_id IS DISTINCT FROM v_membership_id) THEN
                RAISE EXCEPTION USING ERRCODE = '40001',
                    MESSAGE = 'IAM invitation membership is stale';
            END IF;
        ELSE
            IF p_membership_id IS NULL THEN
                RAISE EXCEPTION USING ERRCODE = '22023',
                    MESSAGE = 'IAM invitation membership identity is required';
            END IF;
            v_membership_id := p_membership_id;
            INSERT INTO iam_organization_memberships (
                id, organization_id, user_id, membership_kind, state,
                revision, created_at, updated_at, tombstoned_at
            ) VALUES (
                v_membership_id, p_organization_id, p_user_id, 'member',
                'active', 1, p_now, p_now, NULL
            );
        END IF;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM invitation role scope is invalid';
    END IF;
    INSERT INTO iam_user_role_assignments (
        id, user_id, role_id, role_scope_kind, role_scope_id,
        organization_id, membership_id, state, revision,
        created_at, updated_at, revoked_at
    ) VALUES (
        p_role_assignment_id, p_user_id, p_role_id,
        p_role_scope_kind, p_role_scope_id, p_organization_id,
        v_membership_id, 'active', 1, p_now, p_now, NULL
    );
    RETURN p_role_assignment_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_grant_role_v1(
    uuid,uuid,uuid,uuid,text,uuid,uuid,timestamptz
) FROM PUBLIC;
